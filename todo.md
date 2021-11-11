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