# Go Audit Logs Example

An example Golang application demonstrating how to use the [WorkOS Golang SDK](https://github.com/workos/workos-go) to send and retrieve Audit Log events. This example is not meant to show a real-world example of an Audit Logs implementation, but rather to show concrete examples of how events can be sent using the Go SDK.

## Go Project Setup

1. Clone the example repository using your preferred secure method (HTTPS or SSH).

   ```bash
   # HTTPS
   git clone https://github.com/workos/go-example-applications.git
   ```

   or

   ```bash
   # SSH
   git clone git@github.com:workos/go-example-applications.git
   ```

2. Navigate to the cloned repository and to the Audit Logs:

   ```bash
   cd go-example-applications/go-audit-logs-example
   ```

3. Obtain and make note of the following values. In the next step, these will be set as environment variables.

   - Your [WorkOS API key](https://dashboard.workos.com/api-keys)
   - Your [WorkOS Client ID](https://dashboard.workos.com/configuration)

4. Create a new file called ".env" in the root of the project and add the following variables, replacing the xxx with the value from step 3:
   - WORKOS_API_KEY=xxx
   - WORKOS_CLIENT_ID=xxx

## Start the server

5. The final setup step is to start the server.

   ```bash
   go run main.go
   ```

   You'll know the server is running when you see no errors in the CLI, and output similar to the following is displayed:

   ```bash
   launching audit logs demo with configuration: {Addr::8000 Domains:}
   ```

   Navigate to `localhost:8000` in your web browser to view the example application. Then, click the buttons to see example events after configuring your audit logs in the WorkOS dashboard.

## Audit Logs Setup with WorkOS

6. Follow the [Audit Logs configuration steps](https://workos.com/docs/audit-logs/emit-an-audit-log-event/sign-in-to-your-workos-dashboard-account-and-configure-audit-log-event-schemas) to set up the following 5 events that are sent with this example:

Action title: "user.signed_in" | Target type: "team"
Action title: "user.logged_out" | Target type: "team"
Action title: "user.organization_set" | Target type: "team"
Action title: "user.organization_deleted" | Target type: "team"
Action title: "user.connection_deleted" | Target type: "team"

7. Next, take note of the Organization ID for the Org which you will be sending the Audit Log events for. This ID gets entered into the splash page of the example application.

8. Once you enter the Organization ID and submit it, you will be brought to the page where you'll be able to send the audit log events that were just configured. You'll also notice that the action of setting the Organization triggered an Audit Log already. Click the buttons to send the respective events.

9. To obtain a CSV of the Audit Log events that were sent for the last 30 days, click the "Export Events" button. This will bring you to a new page where you can download the events. Downloading the events is a 2 step process. First you need to create the report by clicking the "Generate CSV" button. Then click the "Access CSV" button to download a CSV of the Audit Log events for the selected Organization for the past 30 days.

## Need help?

If you get stuck and aren't able to resolve the issue by reading our API reference or tutorials, you can reach out to us at support@workos.com and we'll lend a hand.
