# go-directory-sync-example
An example Go application demonstrating how to use the [WorkOS Go SDK](https://github.com/workos-inc/workos-go) for directory sync.

## Prerequisites
- Go

## Go Project Setup

1. In your CLI, navigate to the directory into which you want to clone this git repo.
   ```bash
   $ cd ~/Desktop/
   ```

2. Clone this git repo using your preferred secure method (HTTPS or SSH).
   ```bash
   # HTTPS
   $ git clone https://github.com/workos-inc/Go-Example-Applications.git
   ```

   or

   ```bash
   # SSH
   $ git clone git@github.com:workos-inc/Go-Example-Applications.git
   ```

3. Navigate to the cloned repo.
   ```bash
   $ cd Go-Example-Applications/go-directory-sync-example
   ```

4. Obtain and make note of the following values. In the next step, these will be set as environment variables.
   - Your [WorkOS API key](https://dashboard.workos.com/api-keys)
   - Your Directory ID


5. Create a file called ".env" in the root of the project and add the following environment variables, replacing xxx with the values from your WorkOS dashboard. 
   - WORKOS_API_KEY=xxx
   - WORKOS_DIRECTORY_ID=xxx

6. The final setup step is to start the server.
   ```bash
   $ go run .
   ```

   You'll know the server is running when you see no errors in the CLI.

   Navigate to `localhost:3042` in your web browser. You should see a "Users" button. If you click this link, you should see your users list.

   You can stop the local server for now by entering `CTRL + c` on the command line.

## Need help?

If you get stuck and aren't able to resolve the issue by reading our API reference or tutorials, you can reach out to us at support@workos.com and we'll lend a hand.
