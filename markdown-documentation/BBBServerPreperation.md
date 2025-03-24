# BBB Server preparation

To install bbb-status you need configured webhooks to bbbstatus.
As covered in this [guide](https://avinyaweb.com/bigbluebutton/bigbluebutton-webhooks-create-new-hook/), you need to
edit your `/usr/local/bigbluebutton/bbb-webhooks/config/default.yml`

```shell
nano /usr/local/bigbluebutton/bbb-webhooks/config/default.yml
```

Find hooks.permanentURLs in your config/default.yml and modify it as below

```yaml
hooks:
  permanentURLs:
    - url: 'https://example.bbbstatus.com/event',
      getRaw: false
```

After that apply the changes to bbb-webhooks:

```shell
service bbb-webhooks restart
redis-cli flushall
```

### Note:

This is a permanent way of installing the webhook, and may result in issues if and when bbbstatus is unavailable.
However this is still the recommended way of installing bbbstatus.