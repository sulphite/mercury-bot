<img align="right" src="https://github.com/sulphite/mercury-bot/assets/49396588/bf866c24-6d71-47af-9036-a9733b30eae5">

# hi! I'm mercury.

a discord gopher that subscribes to RSS feeds and posts updates to your server.

[[ invite link ]](https://discord.com/api/oauth2/authorize?client_id=1175863171479785654&permissions=8&scope=bot)

## commands

`/test` - bot will say hello

## host your own

1. For this you will need to create a new application in the discord developer portal, so you can get a **bot token** and **application id**.
2. Make an invite link and invite the bot into your server.
3. Clone this repo, and in the root folder create a `.env` file. Paste your token and id into this file in this format:
```
TOKEN=your token no quotes
APP_ID=your app id no quotes
```
4. In the terminal, run `go build` to build the executable
5. run the bot with `./mercury-bot`

Hopefully bot is up and running.
