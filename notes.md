https://api.github.com
https://docs.github.com/en/rest/overview/other-authentication-methods#via-oauth-and-personal-access-tokens
https://github.com/notifications
https://docs.github.com/en/rest/reference/activity#notifications
https://docs.github.com/en/developers/apps/getting-started-with-apps/about-apps#about-github-apps
https://docs.github.com/en/developers/apps/getting-started-with-apps/setting-up-your-development-environment-to-create-a-github-app#introduction
https://docs.github.com/en/rest/reference/permissions-required-for-github-apps
https://docs.github.com/en/rest/reference/permissions-required-for-github-apps#permission-on-contents
https://github.com/organizations/spaghettigc/settings/apps/spaghettiapp/advanced
https://mholt.github.io/json-to-go/ 
https://github.com/spaghettigc/spaghetti/settings/access
https://pkg.go.dev/github.com/google/go-github/v39/github#pkg-types
https://pkg.go.dev/fmt#Printf
https://stackoverflow.com/questions/59091824/why-does-printf-leave-an-extra-after-my-output
https://docs.github.com/en/rest/reference/issues#timeline
https://docs.github.com/en/enterprise-server@3.0/developers/webhooks-and-events/events/issue-event-types#review_requested
https://docs.github.com/en/rest/reference/pulls#list-requested-reviewers-for-a-pull-request
https://slack.com/intl/en-gb/help/articles/115005265703-Create-a-bot-for-your-workspace
https://api.slack.com/authentication/best-practices#slack_apps__incoming-webhook-urls
https://api.slack.com/reference/block-kit/blocks
https://github.com/golang/go/wiki/CodeReviewComments#variable-names
https://github.com/rakyll/gotest
https://blog.questionable.services/article/testing-http-handlers-go/
https://code.visualstudio.com/docs/languages/go#_import-packages
https://github.com/dnaeon/go-vcr/blob/dd1bc740014d441c053d3dc9119ba533871c7f0c/v2/recorder/recorder.go#L144
https://github.com/dnaeon/go-vcr/blob/dd1bc740014d441c053d3dc9119ba533871c7f0c/cassette/cassette.go#L94
https://github.com/dnaeon/go-vcr/blob/0a1f2acce90f079b99ed7bd78e5000ae3d05b620/cassette/cassette.go#L192
https://pkg.go.dev/net/http#Request
https://docs.github.com/en/rest/reference/issues#timeline

- teams with rota
  - team + user
- teams without rota
  - team

- Cecile was assigneed. PR Title(reponame/#123)
- BR is requested to review by someone
- body


https://github.com/organizations/spaghettigc/settings/apps/spaghettiapp
https://github.blog/changelog/2021-09-29-new-code-review-assignment-settings-and-team-filtering-improvements/

- Webhook 
  - PR # (issue number)
  - Assigned team
- Individual assigned to me from API? - username

https://pkg.go.dev/golang.org/x/net/html#Parse
https://github.com/PuerkitoBio/goquery
https://www.w3.org/TR/selectors/#attribute-substrings

# notes of 2021-12-09
https://api.github.com/repos/spaghettigc/spaghetti/issues/events/5739922110
https://github.com/organizations/spaghettigc/settings/apps/spaghettiapp/advanced
https://docs.github.com/en/rest/reference/repos#webhooks
318500039
https://github.com/organizations/spaghettigc/settings/apps/spaghettiapp/hooks/318500039/deliveries/6388129845
https://codebeautify.org/string-to-json-online
https://api.github.com/repos/spaghettigc/spaghetti/issues/4/events
https://docs.github.com/en/rest/guides/traversing-with-pagination#changing-the-number-of-items-received
https://github.com/spaghettigc/spaghetti/pull/4#event-5739922132

There is no event id in the webhooks

requested reviewers webhook => pr updated_at => timeline events => review requested event + created_at = pr.update_at => get the event id => hit the event api
verification: verify is the users requester and the reviewers are the same
if the verification check fails + we can't find any matching timeline events, we keep the logs, no notification, deal with it later
slack message is 1/reviewer
1 webhook/events can be multiple slack messages

https://stackoverflow.com/a/40323622

2 events for the same review request: 1 for the team, 1 for the assigned team member
5739922132 (assignee with team from the ui)
5739922110 (team only)

Right now we support 1 team being requested, need to think of how to handle when event has several team requests for with several assignees

We still need to move the GH pagination issue events timeline logic to work with the webhook to infer event ID
- we got assginee = 0 when we make the formatmessage.GetAssignedReviewersAndTeam call, need to figure out why