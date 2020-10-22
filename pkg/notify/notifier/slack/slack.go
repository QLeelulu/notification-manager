package slack

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	json "github.com/json-iterator/go"
	"github.com/kubesphere/notification-manager/pkg/notify/config"
	"github.com/kubesphere/notification-manager/pkg/notify/notifier"
	"github.com/prometheus/alertmanager/template"
	"net/http"

	"time"
)

const (
	DefaultSendTimeout = time.Second * 3
	URL                = "https://slack.com/api/chat.postMessage"
	DefaultTemplate    = `{{ template "nm.default.text" . }}`
)

type Notifier struct {
	notifierCfg  *config.Config
	slack        []*config.Slack
	timeout      time.Duration
	logger       log.Logger
	template     *notifier.Template
	templateName string
}

type slackRequest struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type slackResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func NewSlackNotifier(logger log.Logger, receivers []config.Receiver, notifierCfg *config.Config) notifier.Notifier {

	var path []string
	opts := notifierCfg.ReceiverOpts
	if opts != nil && opts.Global != nil {
		path = opts.Global.TemplateFiles
	}
	tmpl, err := notifier.NewTemplate(path)
	if err != nil {
		_ = level.Error(logger).Log("msg", "SlackNotifier: get template error", "error", err.Error())
		return nil
	}

	n := &Notifier{
		notifierCfg:  notifierCfg,
		timeout:      DefaultSendTimeout,
		logger:       logger,
		template:     tmpl,
		templateName: DefaultTemplate,
	}

	if opts != nil && opts.Slack != nil {

		if opts.Slack.NotificationTimeout != nil {
			n.timeout = time.Second * time.Duration(*opts.Slack.NotificationTimeout)
		}

		if len(opts.Slack.Template) > 0 {
			n.templateName = opts.Slack.Template
		}
	}

	for _, r := range receivers {
		receiver, ok := r.(*config.Slack)
		if !ok || receiver == nil {
			continue
		}

		if receiver.SlackConfig == nil {
			_ = level.Warn(logger).Log("msg", "SlackNotifier: ignore receiver because of empty config")
			continue
		}

		n.slack = append(n.slack, receiver)
	}

	return n
}

func (n *Notifier) Notify(data template.Data) []error {

	var errs []error
	msg, err := n.template.TemlText(n.templateName, n.logger, data)
	if err != nil {
		_ = level.Error(n.logger).Log("msg", "SlackNotifier: generate message error", "error", err.Error())
		return append(errs, err)
	}

	send := func(c *config.Slack) error {
		ctx, cancel := context.WithTimeout(context.Background(), n.timeout)
		defer cancel()

		sr := &slackRequest{
			Channel: c.Channel,
			Text:    msg,
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(sr); err != nil {
			_ = level.Error(n.logger).Log("msg", "SlackNotifier: encode message error", "error", err.Error())
			return err
		}

		request, err := http.NewRequest(http.MethodPost, URL, &buf)
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")

		token, err := n.notifierCfg.GetSecretData(c.GetNamespace(), c.SlackConfig.Token)
		if err != nil {
			_ = level.Error(n.logger).Log("msg", "SlackNotifier: get token secret", "error", err.Error())
			return err
		}

		request.Header.Set("Authorization", "Bearer "+token)

		body, err := notifier.DoHttpRequest(ctx, nil, request.WithContext(ctx))
		if err != nil {
			_ = level.Error(n.logger).Log("msg", "SlackNotifier: do http error", "error", err)
			return err
		}

		var slResp slackResponse
		if err := json.Unmarshal(body, &slResp); err != nil {
			_ = level.Error(n.logger).Log("msg", "SlackNotifier: decode response body error", "error", err)
			return err
		}

		if !slResp.OK {
			_ = level.Error(n.logger).Log("msg", "SlackNotifier: slack error", "error", slResp.Error)
			return fmt.Errorf("%s", slResp.Error)
		}

		_ = level.Debug(n.logger).Log("msg", "SlackNotifier: send message", "to", c.Channel)

		return nil
	}

	for _, s := range n.slack {
		err := send(s)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
