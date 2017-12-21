# slack-bot

`slack-bot` monitors a list of queues passed in an environment variable for the 
queue depth and will output to a slack channel when a queue exceed the maximum depth 
allowed.

##### Environment Variables

- MONITORED_QUEUES = comma separated list of queue names to monitor
- MAX_QUEUE_DEPTHS - a comma separated list of max queue depths allowed (same order as MONITORED_QEUEUS)
- SLACK_CHANNEL - Slack Channel to output alerts on

