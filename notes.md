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

We still need to move the GH pagination issue events timeline logic to work with the webhook to infer event ID => OK
- we got assginee = 0 when we make the formatmessage.GetAssignedReviewersAndTeam call, need to figure out why  => OK

# notes of 16/12/2021

We will pick the lastest event ID, visit the event page with a headless browser as the event is rendered on the client side
and is often missing from the original HTML for long PRs.
We need to look for a headless browser solution for golang.

Headless browsers:
https://github.com/headzoo/surf
https://github.com/sourcegraph/webloop#rendering-static-html-from-a-dynamic-single-page-angularjs-app

Surf no JS support
https://github.com/headzoo/surf/blob/a4a8c16c01dc47ef3a25326d21745806f3e6797a/docs/api/packages/browser.md#func-browser-click

We need macports (homebrew alternative) to install webgtk stuff:
need to be installed beforehand
https://github.com/sourcegraph/webloop/issues/3#issuecomment-376425702

# notes of 13/01/2022
installing webkit3 ain't working
Package javascriptcoregtk-3.0 was not found in pkg-config search path: https://github.com/sourcegraph/webloop/issues/9

if we run 

pkg-config --variable pc_path pkg-config

to find pkg-config, not sure reading https://askubuntu.com/questions/210210/pkg-config-path-environment-variable

temporarily try `export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig`

temporarily giving up on webloop due to build issues, looking for alternatives:
https://github.com/go-rod/rod => this works
https://github.com/chromedp/chromedp => haven't tried

FYI we gave up on webloop cause webgtk3 is not easy to install

We have managed to use rod as a headless browser and get the xx has requested a review from yy (assigned from zz).
We fixed regex to parse the sentence.
Works for 1 requested review from team
Breaks for 2 requested review from 2 teams, because of our assumption to pick the last event id.
We received 3 webhooks and sent out 3 identical messages. We missed 1 user being requested.
Solution 1 - store seen events
 - store the seen event ids somewhere and exclude it from the GetEvent logic
Solution 2 - batching
 - group webhooks by their UpdatedAt to make only one request to ListIssueTimeline(), to figure out how many messages we need to send
 - we can limit the search of the issue timeline by restricting to the number of webhooks that have been grouped (e.g. with 2 reviewers on 2 teams, there's 3 webhooks, so at most attempt to find 3 event IDs we're interested in)
 We voted for solution 2

# notes of 20/01/2022
 Potential batching libs:
 https://pkg.go.dev/gocloud.dev/pubsub/batcher#Options - woops ignore this cause it's inside of pubsub
 https://github.com/MasterOfBinary/gobatch - 19 stars ehhh...
 https://github.com/RashadAnsari/go-batch - 6 stars no thanks
 https://github.com/cheggaaa/mb - 13 stars meeh...
 https://pkg.go.dev/search?q=batch+process&m=package

 Max number of webhooks = number of github GC teams

Solution 1
- More reliable for handling late webhooks and retried webhooks, though it's boring it's probably the right solution
- We store the event id we pick, and updatedAt
- Libraries we looked at:
  -  https://github.com/eko/gocache
Solution 2
- Worst case: we may have related webhooks being distributed to different batches, therefore reselecting the same event ids and missing some event ids.
To overcome this, we'll still need to have a stateful system, which defeats the purpose of going with solution 2.
This may occur for delayed webhooks or retried webhooks

https://go.dev/play/p/1Ghqr70Gw5

1 webhook -> pull all related events -> store eventids -> send the slack message
second webhook -> pull all related events -> stop here as we have seen the event ids

We have successfully solved the issue of missing reviewers with solution 1.
Next steps
- code clean up: take the webhook, store in cache and complete HTTP handler so github doesn't think we timed out, separately fetch from cache then process
- test once we have divided into smaller chunks
- we may timeout github webhook as we keep it until we have processed everything
- explore if cache access is threadsafe, else there's a possibility of two webhooks causing the same message to be sent multiple times

# notes of 27/01/2022

- store PR URL as cache value and eventID as cache key

https://github.com/eko/gocache#a-marshaler-wrapper - gocache has marshaller to wrap cacheManager for convenience if bigcache return values as a string

We need to support request 1 team (without rota)
e.g. "ThePesta requested a review from spaghettigc/betterspaghettiteam"
- We have refactor the code and separated some logic (e.g. message, github)
- We now make an async call to send the slack message
- We need to validate if our refactoring is correct or not (working + right separation), how easy it is to write test for the current structure.
- Add logging libraries, error handling, metrics, cli
- explore if cache access is threadsafe, else there's a possibility of two webhooks causing the same message to be sent multiple times
- Make our repo private and make an authenticated call to our PR to scrape the event ID
- Early return if we found the eventIDs instead of waiting for 1s time.Sleep(1 * time.Second) - https://go-rod.github.io/#/selectors/README?id=race-selectors

# notes of 03/02/2022
https://github.com/spaghettigc/spaghetti/timeline_focused_item?after_cursor=Y3Vyc29yOnYyOpPPAAABfFpVwXABqjU0Mjc1ODgyNjc=&before_cursor=Y3Vyc29yOnYyOpPPAAABfptESjgBqjU5NjMwMDY3MDU=&id=PR_kwDOGB7j384scVH7&anchor=event-5603690621


https://github.com/spaghettigc/spaghetti/timeline_focused_item?after_cursor=Y3Vyc29yOnYyOpPPAAABfr8voggBqjYwMDE4MDI1MTY%3D&id=PR_kwDOGB7j384yBHkV&anchor=event-6001802516

spaghettigc/spaghetti/timeline_focused_item?after_cursor=Y3Vyc29yOnYyOpPPAAABfr8voggBqjYwMDE4MDI1MTY%3D&id=PR_kwDOGB7j384yBHkV


#js-timeline-progressive-loader

data-timeline-item-src="spaghettigc/spaghetti/timeline_focused_item?after_cursor=Y3Vyc29yOnYyOpPPAAABfr8voggBqjYwMDE4MDI1MTY%3D&id=PR_kwDOGB7j384yBHkV"

https://github.com/spaghettigc/spaghetti/timeline_focused_item?after_cursor=Y3Vyc29yOnYyOpPPAAABfr9ETwABqjYwMDE5NDQ0MjA%3D&id=PR_kwDOGB7j384yBHkV&anchor=event-6001944400

https://github.com/spaghettigc/spaghetti/timeline_focused_item?after_cursor=Y3Vyc29yOnYyOpPPAAABfr9ETwABqjYwMDE5NDQ0MjA%3D&id=PR_kwDOGB7j384yBHkV&anchor=event-6001944420

https://github.com/spaghettigc/spaghetti/timeline_focused_item?after_cursor=Y3Vyc29yOnYyOpPPAAABfFpVwXABqjU0Mjc1ODgyNjc%3D&before_cursor=Y3Vyc29yOnYyOpPPAAABfptESjgBqjU5NjMwMDY3MDU%3D&id=PR_kwDOGB7j384scVH7&anchor=event-5427505920


We created a new PR and tested our project against it, and it failed when trying to retrieve text displayed in the UI for particular event IDs. This is because the front end has some smart ass logic to not display hidden items if there isn't enough of them. We figured out a workaround by looking at the hidden items loader div, which makes a request to /timeline_focused_item with some cursor info alongside the event ID. We have no clue how to reverse engineer the cursor information, but since it's already available on the div with id `js-timeline-progressive-loader`, we can just rebuild the request that would of been made if there was enough timeline items that are hidden. We tested this and it seems to work fine.
1. Find the event ids
2. Visit the event id page
3. Find div id `js-timeline-progressive-loader`
4. Grab the `data-timeline-item-src`
5. Make a request to `https://github.com/` + `{data-timeline-item-src}` + `&anchor=event-` + `{eventId}`
6. Continue as usual

We have implemented this solution and it worked.

We are trying to support requesting a team only, without rota. We've noticed that rod currently only return `GitCecile requested a review from Feb 3, 2022`, without the team name. Turns out that the team name will be displayed only when authenticated. We had a closer look at the authentication on the browser, github reads a session token stored in the cookies, this means that we might not be able to use an access token and pass it in the header, we may need to authenticate at each request.
It seems like rod only support username/password authentication at a first glance. https://pkg.go.dev/github.com/go-rod/rod@v0.101.8#Browser.HandleAuth

https://docs.github.com/en/rest/guides/basics-of-authentication

We had the idea of using `data-hovercard-type` to distinguish between a team being requested without rota and a team being requested with rota, however it seems like this is also used for the requester, we'll always get `data-hovercard-type="user"` as a user has requested a review.
data-hovercard-type="team"
data-hovercard-type="user"

# notes of 17/02/2022
https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-a-github-app
https://github.com/google/go-github/blob/v17.0.0/github/github.go#L300

We tried using JWT token shenanigans to access /timeline_focused_item endpoint, but it doesn't work as we discovered it uses a user session cookie, we're going to use the normal GH login from a web browser.

We checked the peak PR created per hour with look https://looker.gocardless.io/sql/t4qk2822bvttqn, at most was ~200 an hour. Github has a user to server rate limit of 5k requests/h, so we should be fine for a long time, given that we're assuming one PR = one team request, which is one if there's no github rota, two if they have a rota.
We have decided to log in to github via the headless browser, maintain the session and visit the timeline focused item page, this solution works.

So we are now able to process review request for a team without rota, however, we need to work on the regex to parse the message correctly (no slack message).
Another question will be, how to distinguish different types of requests.

We have triggered 2 different types of requests (with/without rota) and it works, we have visited the right timeline focused items.

[] Add regex for request review from 1 team without rota 
[] Review the console prints `fmt.Printf`, we should be using real loggings

# notes of 24/02/2022
We want to support request review from team with rota only for now, we prefer to productionise first over adding new features (e.g. support request review from 1 team without rota)
[] Review the console prints `fmt.Printf`, we should be using real loggings
[x] Error handling cause we're panic()-ing everywhere
- We need to validate if our refactoring is correct or not (working + right separation), how easy it is to write test for the current structure.
- Add logging libraries(github.com/uber-go/zap), error handling, metrics, cli
- explore if cache access is threadsafe, else there's a possibility of two webhooks causing the same message to be sent multiple times
- Make our repo private and make an authenticated call to our PR to scrape the event ID
- Dockerize
- Deploy to Heroku/Consider making it a utopia app now
- Add a readme with setup instructions (which account should devs use for headless browser github login?)


https://github.com/gocardless/cookiecutter-georges/pull/26 - Switch in Georges cookie cutter logging from zerolog to zap https://gocardless.slack.com/archives/C03H92E3B/p1637674549024000?thread_ts=1631114597.052900&cid=C03H92E3B

https://github.com/gocardless/cookiecutter-georges/pull/26/files?file-filters%5B%5D=.go&show-viewed-files=true#diff-e23226f59a690fb97dce9ec4d14fb268e41b01e529f1b77eec2a3abb014ba00fR81

# notes of 17/03/2022
- After turning on private repo, we can't request teams anymore https://github.com/pricing
- We want to productionise first
  - [x]logging - Jason
  - [x] makefile
  - [x]readme setup - Cecile -> Initiated a readme that needs to be refined
  - [x] error handling
  - [] tests
- Deploy to Heroku
- Dockerize
- Make the repo private
- Move to utopia

https://go-rod.github.io/#/context-and-timeout?id=cancellation

# notes of 31/03/2022

- Add golangci-lint later
https://github.com/gocardless/billing/blob/attempt-1/cmd/billing/main.go
https://github.com/gocardless/billing/pull/80
https://github.com/gocardless/cookiecutter-georges/blob/aa3a499cf46584091d2296b88295cec94a271b3b/cmd/app/main.go#L122\]
https://docs.sentry.io/product/sentry-basics/integrate-frontend/create-new-project/
https://github.com/gocardless/dispatcher/search?q=captureException
https://blog.sentry.io/2019/08/15/introducing-sentrys-unified-go-sdk#capturing-errors

- We added sentry for error handling but the error stack trace is not working properly.
https://blog.sentry.io/2019/08/15/introducing-sentrys-unified-go-sdk#capturing-errors We tried this example and it works but the stack doesn't show up for our own main function.



# notes of 07/04/2022
After reading this https://incident.io/blog/golang-errors
We found that stdlib errors doesn't give stack traces at all, we have to use pkg/errors combined with errors.Wrap()
This reflects the same duplicate stacktrace problem in the article above, but it works for our usecase to see the stack trace from whereever we call return errors.Wrap(err, "some msg")

https://github.com/gocardless/gc-cli/blob/bab65d2e4b80fdc624e27671759f0776c78ef578/helper/sentry.go#L32
https://pkg.go.dev/github.com/getsentry/sentry-go?utm_source=godoc#Hub
https://pkg.go.dev/github.com/getsentry/sentry-go?utm_source=godoc#Scope
https://pkg.go.dev/github.com/getsentry/sentry-go?utm_source=godoc#Breadcrumb
https://docs.sentry.io/product/issues/issue-details/breadcrumbs/

We have refactored error handling with errors.Wrap.

# notes of 14/04/2022

https://github.com/rakyll/gotest
https://github.com/eko/gocache/blob/master/cache/cache_test.go
https://github.com/eko/gocache#write-your-own-custom-cache
https://pkg.go.dev/go.uber.org/zap@v1.21.0/zaptest#NewLogger
https://github.com/eko/gocache/blob/master/test/mocks/cache/cache_interface.go

We need to review the condition in lookup
```
if err != nil && err != bigcache.ErrEntryNotFound {}
```

We have managed to control the behaviour of the cache by mocking the store.
```
	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Get("event1").Return("event1", nil)
	store.EXPECT().Get("event2").Return(nil, )

	cache := cache.New(store)
```

# notes of 28/04/2022
https://github.com/gocardless/anu/blob/master/terraform/github/users/users.jsonnet
https://onsi.github.io/gomega/#making-assertions
https://github.com/eko/gocache/blob/1a9de67e673d64a61821c0fe3bc8e1548a7333f1/cache/cache_test.go#L32-L55
https://github.com/eko/gocache/blob/1a9de67e673d64a61821c0fe3bc8e1548a7333f1/marshaler/marshaler_test.go#L126-L149

# notes of 12/05/2022
https://github.com/spaghettigc/spaghetti/commit/f7b6c27d47ed8cd677a34878e88371ec97e741c6#diff-9ab20159ff898156229b365589a222ca8277c74f0a26464481f60bf12b2d5d0bL36

We need to finish the tests in Message
- done for Truncate and FormatMessage