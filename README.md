# Spaghetti APP
Slack app sending notifications when teams are requested for Github pull request reviews.


# Getting started
## 1. Install dependencies
```
make install
```

## 2. Setup your integration
### 2.1. Slack
Follow the [create an app](https://api.slack.com/authentication/basics#creating) tutorial.

### 2.2 Github
#### App
[Create a github app](https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app)
#### Robot account
[Create a robot github account] (https://github.com/signup)

## 3. Environment variables

## 4. Start the server
For local development, you can host the server with [ngrok](https://ngrok.com/), you'll need to update the Webhook URL of your Github app.
Run the following to start the server:
```
make
```

