## The Spec

### High level

The goal of project is to create a server that acts as a simple authenticating reverse-proxy. At a high level, the flow is:

- User sends their username and password as a BasicAuth request. If valid, a HMAC token is returned. Invalid requests increment a counter.
- User presents the token generated in the previous step as an [authorization bearer token](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization) to authenticate against the reverse proxy endpoints.
- If authenticated, the user's requests should be proxied to an allowed destination route.

### Steps

1. Create an application that loads the included configuration file (`./config.json)`. That configuration file contains:

  - A collection of usernames and bcrypt password hashes. The corresponding clear-text passwords for those users are:

    - bob: `hunter2`
    - alice: `password`
    - eve: `qwerty`

  - A collection of URLs and associated allowed users to be reverse proxied.

  - A base64 encoded 256 bit key to be used to sign the authorization tokens.

2. Add a HTTP endpoint (`/session/create`) that accepts a `GET` request with [basic authentication](https://en.wikipedia.org/wiki/Basic_access_authentication). If the basic auth credentials match the bcrypt hash, return a signed token. The endpoint should:

  - return `404` if the username is not found.
  - return `403` if the username exists, but the password's _hash_ does not match. Increment the `unauthorized_attempt` stats counter.
  - return `200` if the username/password hash matches and reply with the user token in the response body. Also, increment the `authorized_attempt` counter, and update the `last_successful_session_unix_time` counter. The signed token should contain (ok to wrap in base64):
  - _username_
  - _issued timestamp_
  - _expiry timestamp_
  - SHA256 HMAC of the above values.

3. Add a HTTP endpoint (`/session/{username}/stats`) that accepts a `GET` request. It should return a JSON response that returns the user specific stat counters from the previous step.

4. Add a HTTP path (`/proxy/{name}`) where `{name}` corresponds to each of config file's `allowed_routes.name` values and `allowed_routes.destination` is where the reverse proxy's target location is. This handler should look for a valid token from the previous step in the [authorization bearer](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization) header. The handler should:

  - return `404` if the route `{name}` is not known.
  - return `401` if the route `{name}` is known, but token hash is invalid or empty.
  - return `401` if the route `{name}` is known, the token hash is valid, but the token is not yet, or no longer valid.
  - return `403` if the route `{name}` is known, the token hash is valid, but the _username_ does not have access to this particular route.
  - Otherwise, if the requesting _username_ matches an `allowed_routes.users` entry, allow and proxy the request.
