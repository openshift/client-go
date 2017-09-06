package client

import (
	oauthapi "github.com/openshift/origin/pkg/oauth/apis/oauth"
)

type OAuthAuthorizeTokensInterface interface {
	OAuthAuthorizeTokens() OAuthAuthorizeTokenInterface
}

type OAuthAuthorizeTokenInterface interface {
	Create(token *oauthapi.OAuthAuthorizeToken) (*oauthapi.OAuthAuthorizeToken, error)
	Delete(name string) error
}

type oauthAuthorizeTokenInterface struct {
	r *Client
}

func newOAuthAuthorizeTokens(c *Client) *oauthAuthorizeTokenInterface {
	return &oauthAuthorizeTokenInterface{
		r: c,
	}
}

func (c *oauthAuthorizeTokenInterface) Delete(name string) (err error) {
	err = c.r.Delete().Resource("oauthauthorizetokens").Name(name).Do().Error()
	return
}

func (c *oauthAuthorizeTokenInterface) Create(token *oauthapi.OAuthAuthorizeToken) (result *oauthapi.OAuthAuthorizeToken, err error) {
	result = &oauthapi.OAuthAuthorizeToken{}
	err = c.r.Post().Resource("oauthauthorizetokens").Body(token).Do().Into(result)
	return
}
