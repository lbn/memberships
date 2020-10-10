package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/dgrijalva/jwt-go"
)

var (
	bearerExp = regexp.MustCompile("^Bearer (.*)$")
	issuer    = os.Getenv("auth_issuer")
	audience  = os.Getenv("auth_audience")
)

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(fmt.Sprintf("%s.well-known/jwks.json", issuer))

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}

func getToken(event events.APIGatewayCustomAuthorizerRequest) (string, error) {
	if event.Type != "TOKEN" {
		return "", errors.New(`Expected "event.type" parameter to have value "TOKEN"`)
	}
	if event.AuthorizationToken == "" {
		return "", errors.New(`Expected "event.authorizationToken" parameter to be set`)
	}
	match := bearerExp.FindStringSubmatch(event.AuthorizationToken)
	if len(match) < 2 {
		return "", fmt.Errorf(`Invalid Authorization token - %s does not match "Bearer .*"`, event.AuthorizationToken)
	}
	return match[1], nil
}

// Help function to generate an IAM policy
func generatePolicy(principalId, effect, resource string) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}

	authResponse.Context = map[string]interface{}{}
	return authResponse
}

func HandleRequestAuthorize(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (res events.APIGatewayCustomAuthorizerResponse, err error) {
	token, err := getToken(event)
	if err != nil {
		return
	}

	parsedToken, err := jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		claims := jwtToken.Claims.(jwt.MapClaims)
		checkAud := claims.VerifyAudience(audience, false)
		if !checkAud {
			return nil, fmt.Errorf("invalid audience - want %s but got %s", audience, claims["aud"])
		}
		// Verify 'iss' claim
		checkIss := claims.VerifyIssuer(issuer, false)
		if !checkIss {
			return nil, fmt.Errorf("invalid issuer - want %s but got %s", issuer, claims["iss"])
		}

		cert, err := getPemCert(jwtToken)
		if err != nil {
			return nil, err
		}

		return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
	})
	if err != nil {
		err = err.(*jwt.ValidationError).Inner
		return
	}
	principalID, ok := parsedToken.Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err = errors.New("invalid token")
		return
	}
	return generatePolicy(principalID, "Allow", event.MethodArn), nil
}
