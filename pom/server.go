package pom

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

// NewServer returns a configured server with its required dependencies
func NewServer(listenAddress string, users UserStore, routes RouteStore, statsKeeper UserStats, auth AuthProvider) *PomServer {
	return &PomServer{
		address:      listenAddress,
		userStore:    users,
		routeStore:   routes,
		userStats:    statsKeeper,
		authProvider: auth,

		proxies: make(map[string]httputil.ReverseProxy),
	}
}

// UserStore is expected to provide password checks on users
type UserStore interface {
	CheckUsername(string) bool
	CheckPassword(string, string) bool
}

// RouteStore is expected to provide routes and the permissions to those routes
type RouteStore interface {
	IsUserAllowed(string, string) bool
	GetAddress(string) (string, bool)
}

// UserStats keeps track of a user's login attempts
type UserStats interface {
	IncrementAuthorizedAttempt(string)
	IncrementUnauthorizedAttempt(string)
	GetStats(string) string // should return a json body
}

// AuthProvider generates and validates pom auth tokens
type AuthProvider interface {
	GetToken(string) (string, error)
	ValidateToken(string) (string, bool)
}

// PomServer is a reverse-proxy authentication server
type PomServer struct {
	// Injected dependencies
	address      string
	userStore    UserStore
	routeStore   RouteStore
	userStats    UserStats
	authProvider AuthProvider

	// Caching
	proxies map[string]httputil.ReverseProxy
}

// ListenAndServe runs the PomServer, handling incoming requests
func (me *PomServer) ListenAndServe() error {
	mux := mux.NewRouter()
	me.registerRoutes(mux)
	if err := http.ListenAndServe(me.address, mux); err != nil {
		return err
	}
	return nil
}

func (me *PomServer) registerRoutes(mux *mux.Router) {
	mux.HandleFunc("/session/create", me.HandleSessionCreate).Methods("GET")
	mux.HandleFunc("/session/{user}/stats", me.AuthMiddleware(me.HandleStatsRequest)).Methods("GET")
	mux.HandleFunc("/proxy/{target}", me.AuthMiddleware(me.HandleProxyRequest))
}

func (me *PomServer) HandleSessionCreate(res http.ResponseWriter, req *http.Request) {
	// read username & password from auth header
	username, password, hasAuth := req.BasicAuth()
	if !hasAuth {
		log.Printf("Create session: No authentication provided")
		res.Header().Set(`WWW-Authenticate`, `Basic realm="Restricted"`)
		http.Error(res, "Not authenticated", 401)
		return
	}

	userFound := me.userStore.CheckUsername(username)
	if !userFound {
		log.Printf("Create session: Username not found")
		http.Error(res, "Username not found", 404)
		return
	}

	authenticated := me.userStore.CheckPassword(username, password)
	if !authenticated {
		log.Printf("Create session: invalid password")
		http.Error(res, "Forbidden", 403)
		me.userStats.IncrementUnauthorizedAttempt(username)
		return
	}

	// Return JWT authorization token
	token, err := me.authProvider.GetToken(username)
	if err != nil {
		log.Printf("Create session: failed to get auth token for user %s: %s", username, err.Error())
		http.Error(res, "Internal err", 500)
		return
	}
	me.userStats.IncrementAuthorizedAttempt(username)
	io.WriteString(res, fmt.Sprintf(`{"token":"%s"}`, token))
}

func (me *PomServer) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authorization")
		authSplit := strings.Split(auth, " ")
		if len(authSplit) != 2 {
			log.Printf("Auth middleware: bad auth header")
			http.Error(res, "Not authorized", 401)
			return
		}

		authType, token := authSplit[0], authSplit[1]
		if strings.ToLower(authType) != "bearer" {
			log.Printf("Auth middleware: auth token not Bearer")
			http.Error(res, "Not authorized", 401)
			return
		}

		username, valid := me.authProvider.ValidateToken(token)
		if !valid {
			log.Printf("Auth middleware: invalid token")
			http.Error(res, "Not authorized", 401)
			return
		}

		vars := mux.Vars(req)
		if vars == nil {
			log.Printf("Auth middleware: no vars on request")
			http.Error(res, "Internal error", 500)
			return
		}
		vars["username"] = username

		// User is good to go, serve!
		log.Printf("Auth middleware: success for username=%s", username)
		next(res, req)
	})
}

func (me *PomServer) HandleStatsRequest(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	// Ensure user is requesting the stats for themselves
	username, ok := params["username"]
	if !ok {
		log.Printf("Stats request: username not found")
		http.Error(res, "Internal Error", 500)
		return
	}
	user, ok := params["user"]
	if !ok {
		log.Printf("Stats request: user not found")
		http.Error(res, "Internal Error", 500)
		return
	}
	if username != user {
		log.Printf("Stats request: username not equal to requested user")
		http.Error(res, "Forbidden", 403)
		return
	}

	// Otherwise, return their stats
	stats := me.userStats.GetStats(user)
	io.WriteString(res, stats)
}

func (me *PomServer) HandleProxyRequest(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	// Ensure user is requesting the stats for themselves
	username, ok := params["username"]
	if !ok {
		log.Printf("Proxy request: username not found")
		http.Error(res, "Internal Error", 500)
		return
	}

	// Get target from url path
	targetName, targetFound := params["target"]
	if !targetFound {
		log.Printf("Proxy request: target not found in path")
		http.Error(res, "Internal error", 500)
		return
	}

	// Check for allowed user from route store
	authorized := me.routeStore.IsUserAllowed(targetName, username)
	if !authorized {
		log.Printf("Proxy request: user not allowed")
		http.Error(res, "Forbidden", 403)
		return
	}

	// Get redirect URL
	target, found := me.routeStore.GetAddress(targetName)
	if !found {
		log.Printf("Proxy request: target address not found for target=%s", targetName)
		http.Error(res, "Target not found", 404)
		return
	}

	url, err := url.Parse(target)
	if err != nil {
		log.Printf("failed to parse url target `%s`", target)
		http.Error(res, "Internal error", 500)
		return
	}

	// We're good to go, so redirect!
	me.updateForRedirection(url, req)
	revProxy := me.getReverseProxy(url)
	revProxy.ServeHTTP(res, req)
}

func (me *PomServer) getReverseProxy(url *url.URL) *httputil.ReverseProxy {
	proxy, ok := me.proxies[url.String()]
	if !ok {
		proxy = *httputil.NewSingleHostReverseProxy(url)
		me.proxies[url.String()] = proxy
	}
	return &proxy
}

func (me *PomServer) updateForRedirection(url *url.URL, req *http.Request) {
	req.Host = url.Host
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.URL.Path = "/"
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
}
