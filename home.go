package main

func Home(token string) string {
	return `<!DOCTYPE html>
<html lang="en">
    <head>
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <script src="https://unpkg.com/htmx.org@1.9.10"></script>
        <script src="https://unpkg.com/htmx.org/dist/ext/ws.js"></script>
        <title>home</title>
        <script>
            document.addEventListener("htmx:configRequest", (event) => {
                event.detail.headers["Authorization"] = 'Bearer ' + localStorage.getItem('key');
            });
            document.addEventListener("htmx:wsConfigSend", (event) => {
                event.detail.headers["Authorization"] = 'Bearer ' + localStorage.getItem('key');
            })
        </script>
    </head>
    <body>
        <div>
            <div hx-post="/user/profile" hx-trigger="load"></div>
        </div>
        <div hx-ext="ws" ws-connect="/socket?tkn=` + token + `">
            <div id="msg"></div>
            <form id="form" ws-send>
                <input type="text" name="text">
            </form>
        </div>
    </body>
</html>`

}
