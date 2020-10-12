# Background

This was a take-home project for writing a small reverse-proxy HTTP service in go.

## Basic Spec
For the full spec prompt given to me, please see the SPEC.md file.

* Create a login endpoint
* Create a stats endpoint
* Create a proxy endpoint

## Limitations
Ultimately, this was a time-limited project, and I did not get done everything I would have liked to, mostly in the testing area.

* 4 hours of setup
* 4 hours of with-prompt building
* Manual end-to-end testing with insomnia

# Instructions

## Build
`make build`

## Run
In order to run pom-server, you must first have a valid .env file to source. See `example_env` for details.  
For this project, you can simply call:
`cp example_env .env`

Then:
`make run`  
or  
`./pom-server`

### Test
Ensure you have a `.env` file, then:  
`make test`

