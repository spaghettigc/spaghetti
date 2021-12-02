[X] check timeline
[X] check webhook for request teams
[x] slack notification for team requested review
[x] slack integration - test connection to API via SDK

[] Mentions on PR comments doesn't work => no webhooks
[] slack notification for team mention in comment,  we need to check if we receive webhooks for mentions, Pull request review comment permission not working atm
[] slack notification for team mention in description


- @billing_and_revenue team was requested for PR#123 title in repository_name by @requester, @team_member was assigned
  - partial body description (nice to have)

[] Code clean up (reorganise) + repo
[] conditional logic to mention the team and or user (team member in rota) for PR review
[] Slack message formatting, quote the description

[] v1: have a static user mapping, stored in code
[] Move the code to gocardless (CI+Sentry/Kibana) + remap user mapping (slack IDs are different between workspaces)
[] v2: support manually entered slack user github user mapping

[] add tests for postmessage and formatmessage
[x] fix makefile

[x] GetAssignee with multiple requested reviewers (we need a third account)

[] try to get the timeline event id
[] fix the code to
  - Assess if the requested review is in the team
  - We want the review that triggers the webhook (timeline events)
[] make the recording code more generic, accept a directory, filename automated from test name

We've tried playing around with the GraphQL explorer with pull requests & reviewers but didn't find anything more than what the REST API gave us back

We thought of two directions to go in:
1) Sacrifice being able to identify the assigned user from a team, tagging the entire team in slack instead
2) Explore Github PR assignee's more, because I have a hunch that the reviewers on a PR is the same as assignees on PRs/issues

The context for 2) is we have no f-ing clue how 
```
ThePesta requested review from spaghettigc/betterspaghettiteam, spaghettigc/obviousspaghettiteam, wmytbaow (assigned from spaghettigc/obviousspaghettiteam)
```
how the information about which user is assigned comes from, because the event https://api.github.com/repos/spaghettigc/spaghetti/issues/events/5672461546 only contains the review requester and requested team.
we tried hitting the timeline API with the golang github client but saw the same stuff as the webhook events.

========
- we need to figure out how to find event id in webhooks
- we got assginee = 0 when we make the formatmessage.GetAssignedReviewersAndTeam call, need to figure out why