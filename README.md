# user-api

## You will need:

 - make   
 - docker   
 - go v1.19+   
 - k6 (for integration tests)

## Quickstart:

To run the project locally from this directory:

Start a dev postgres db in docker:

      make docker-run-postgres-dev

Start the user-api server:

      make run

Or in a container if you prefer:

      make docker-run

Other useful commands:

  psql terminal in dev:

      make docker-run-psql-dev

  run tests:

      make docker-test-integration

Environment:

The app requires the following environment variables to be set:

      POSTGRES_USER
      POSTGRES_PASSWORD
      POSTGRES_URL e.g 0.0.0.0:5432
      POSTGRES_DB

### API:
Implements a http REST API for a users resource. Records are stored in Postgres with the following schema:

      id
      first_name
      last_name
      nickname
      password
      email
      country
      created_at
      updated_at


It has the following endpoints:

**/ GET**
  Healthcheck for service. 
  returns status 200 on success

**/user POST**
  Create a new user. key value pairs are passed using x-www-form-urlencoded.
  
  Available keys:

      first_name
      last_name
      nickname
      password
      email
      country

  returns status 201 on success

**/users GET**
  Get a list of all users. Filter by query params e.g. `
  /users?country=USA&first_name=Hulk` would return all users from the USA with the first name Hulk.
  
  Available query params:

      first_name
      last_name
      nickname
      password
      email
      country

  returns status 200 on success

**/user/:id PATCH**
  Patch an existing user. Key value pairs passed using x-www-form-urlencoded
  
  Available keys:

      first_name
      last_name
      nickname
      password
      email
      country

  returns status 200 on success

**/user/:id DELETE**
  Delete an existing user.

  returns status 200 on success

Available content types:

      application/json

Example deployment available at: https://user-api.i0n.io
