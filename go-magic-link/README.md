# Go-admin-portal-example
An example Golang application demonstrating how to use the [WorkOS Golang SDK](https://github.com/workos-inc/workos-go) so your customers can access the WorkOS Admin Portal from your application.


## Go Project Setup

1. Clone the example repo using your preferred secure method (HTTPS or SSH).
   ```bash
   # HTTPS
   $ git clone https://github.com/workos-inc/Go-Example-Applications.git
   ```

   or

   ```bash
   # SSH
   $ git clone git@github.com:workos-inc/Go-Example-Applications.git
   ```

2. Navigate to the cloned repo and to the admin portal:
   ```bash
   $ cd Go-Example-Applications/go-magic-link
   ```

3. Obtain and make note of the following values. In the next step, these will be set as environment variables.
   - Your [WorkOS API key](https://dashboard.workos.com/api-keys)

4. Add this to your APIKey variable.

## Start the server

5. The final setup step is to start the server.
   ```bash
   go run main.go
   ```

   You'll know the server is running when you see no errors in the CLI, and output similar to the following is displayed:

   ```bash
   launching admin portal demo with configuration: {Addr::8000 Domains:}
   ```
   
   Set the default redirect URI as "http://localhost:8000/success" in your WorkOS Dashboard Configuration tab and navigate to `localhost:8000` in your web browser to view the passwordless example app. 
   
   Hooray!

## Need help?

If you get stuck and aren't able to resolve the issue by reading our API reference or tutorials, you can reach out to us at support@workos.com and we'll lend a hand.
