package main

/*func TestRun(t *testing.T) {

	// Setup real server
	// TODO: make a way to control start/stop
	errChan := make(chan (error))
	go func() {
		defer close(errChan)
		err := run()
		if err != nil {
			errChan <- err
		}
	}()

	//Get basic client
	client := http.Client{}

	// Make the calls
	req, err := http.NewRequest("GET", "http://localhost:8080/session/create", nil)
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth("bob", "hunter2")

	res, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	// Read bearer token from response
	_ = res

	// use token to call get stats

	// use token to call proxy

	// Check for errors from err chan
}*/
