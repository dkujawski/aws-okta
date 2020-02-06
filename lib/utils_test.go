package lib

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/segmentio/aws-okta/lib/saml"
	"github.com/stretchr/testify/assert"
)

var testSAMLHTML = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta name="robots" content="noarchive"/>
		<link href="https://ok3static.oktacdn.com/assets/img.png" rel="icon" type="image/png" sizes="16x16"/>
		<title>Signing in...</title>
		<!--[if IE]><link href="https://ok3static.oktacdn.com/assets/css.css" type="text/css" rel="stylesheet"/><![endif]-->
		<!--[if gte IE 9]><link href="https://ok3static.oktacdn.com/assets/css.css" type="text/css" rel="stylesheet"/><![endif]-->
		<script>if (typeof module === 'object') {window.module = module; module = undefined;}</script>
	</head>	
	<body id="app" class="enduser-app  okta-legacy-theme  ">
		<div id="container">
			<link href="https://ok3static.oktacdn.com/assets/css.css" type="text/css" rel="stylesheet"/>
			<script>var interstitialMinWaitTime = 1200;</script>
	        <div id="okta-interstitial-wrap">
	            <div class="okta-auth-mask-new-interstitial"></div>
		        <div class="new-interstitial" id="new-interstitial">
			        <div id="okta-auth-band">
			            <img src="https://ok3static.oktacdn.com/assets/img.png" width="376" height="160" alt="Please wait" class="new-img-static"/>
			            <h1 id="okta-auth-heading" class ='signing-in-text'>Signing in to AWS</h1>
			   		</div>
		        </div> <!--new-interstitial -->
		    </div> <!--okta-interstitial-wrap -->
		    <form id="appForm" action="https&#x3a;&#x2f;&#x2f;signin.aws.amazon.com&#x2f;saml" method="POST">
		    	<input name="SAMLResponse" type="hidden" value="{{.SAMLValue}}"/>
		    	<input name="RelayState" type="hidden" value=""/>
			</form>
		</div>
	</body>
</html>
`

var testSAMLResponseValueXML = `
<?xml version="1.0" encoding="UTF-8"?>
<saml2p:Response Destination="{{.Dest}}" 
			     ID="{{.SAMLResponseID}}" 
			     IssueInstant="{{.IssueInstant}}" 
			     Version="{{.Version}}" 
			     xmlns:saml2p="urn:oasis:names:tc:SAML:2.0:protocol" 
			     xmlns:xs="http://www.w3.org/2001/XMLSchema">
	<saml2:Issuer Format="urn:oasis:names:tc:SAML:2.0:nameid-format:entity" 
		  	      xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">{{.OKTAID}}</saml2:Issuer>
	<ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
		<ds:SignedInfo>
			<ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
			<ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
			<ds:Reference URI="#{{.SAMLResponseID}}">
				<ds:Transforms>
					<ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
					<ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#">
						<ec:InclusiveNamespaces PrefixList="xs" xmlns:ec="http://www.w3.org/2001/10/xml-exc-c14n#"/>
					</ds:Transform>
				</ds:Transforms>
				<ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
				<ds:DigestValue>testDigestValue</ds:DigestValue>
			</ds:Reference>
		</ds:SignedInfo>
		<ds:SignatureValue>testData</ds:SignatureValue>
		<ds:KeyInfo><ds:X509Data><ds:X509Certificate>testData</ds:X509Certificate></ds:X509Data></ds:KeyInfo>
	</ds:Signature>
	<saml2p:Status xmlns:saml2p="urn:oasis:names:tc:SAML:2.0:protocol"><saml2p:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/></saml2p:Status>
	<saml2:Assertion ID="{{.SAMLAssertionID}}" 
				     IssueInstant="{{.IssueInstant}}" 
				     Version="2.0" 
				     xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion" 
				     xmlns:xs="http://www.w3.org/2001/XMLSchema">
		<saml2:Issuer Format="urn:oasis:names:tc:SAML:2.0:nameid-format:entity" xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
			{{.OKTAID}}
		</saml2:Issuer>
		<ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
			<ds:SignedInfo>
				<ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
				<ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
				<ds:Reference URI="#{{.SAMLAssertionID}}">
					<ds:Transforms>
						<ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
						<ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#">
							<ec:InclusiveNamespaces PrefixList="xs" xmlns:ec="http://www.w3.org/2001/10/xml-exc-c14n#"/>
						</ds:Transform>
					</ds:Transforms>
					<ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
					<ds:DigestValue>testDigestValue</ds:DigestValue>
				</ds:Reference>
			</ds:SignedInfo>
			<ds:SignatureValue>testSignatureValue</ds:SignatureValue>
			<ds:KeyInfo><ds:X509Data><ds:X509Certificate>testCertificateData</ds:X509Certificate></ds:X509Data></ds:KeyInfo>
		</ds:Signature>
		<saml2:Subject xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
			<saml2:NameID Format="urn:oasis:names:tc:SAML:2.0:nameid-format:unspecified">{{.NameID}}</saml2:NameID>
			<saml2:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">
				<saml2:SubjectConfirmationData NotOnOrAfter="{{.ConditionNotOnOrAfter}}" Recipient="https://signin.aws.amazon.com/saml"/>
			</saml2:SubjectConfirmation>
		</saml2:Subject>
		<saml2:Conditions NotBefore="{{.ConditionNotBefore}}" NotOnOrAfter="{{.ConditionNotOnOrAfter}}" xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
			<saml2:AudienceRestriction><saml2:Audience>urn:amazon:webservices</saml2:Audience></saml2:AudienceRestriction>
		</saml2:Conditions>
		<saml2:AuthnStatement AuthnInstant="{{.IssueInstant}}" SessionIndex="{{.SessionIndex}}" xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
			<saml2:AuthnContext><saml2:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport</saml2:AuthnContextClassRef></saml2:AuthnContext>
		</saml2:AuthnStatement>
		<saml2:AttributeStatement xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
			<saml2:Attribute Name="https://aws.amazon.com/SAML/Attributes/Role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:uri">
				<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">
					{{.AWSRole}}
				</saml2:AttributeValue>
			</saml2:Attribute>
			<saml2:Attribute Name="https://aws.amazon.com/SAML/Attributes/RoleSessionName" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic">
				<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">{{.NameID}}</saml2:AttributeValue>
			</saml2:Attribute>
			<saml2:Attribute Name="https://aws.amazon.com/SAML/Attributes/SessionDuration" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic">
				<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">43200</saml2:AttributeValue>
			</saml2:Attribute>
		</saml2:AttributeStatement>
	</saml2:Assertion>
</saml2p:Response>
`

type tplDataForSAMLValue struct {
	Dest                  string
	Version               string
	SAMLResponseID        string
	IssueInstant          string
	OKTAID                string
	SAMLAssertionID       string
	NameID                string
	ConditionNotOnOrAfter string
	ConditionNotBefore    string
	SessionIndex          string
	AWSRole               string
}

type tplDataForSAMLHTML struct {
	SAMLValue string
}

func TestParseSAML(t *testing.T) {
	testTimeNow := time.Now()
	awsID := genRandNumStr(12)
	tps := tplDataForSAMLValue{
		Dest:                  "https://signin.aws.amazon.com/saml",
		Version:               fmt.Sprintf("%1.1f", float32(rand.Intn(5))+rand.Float32()),
		SAMLResponseID:        fmt.Sprintf("id%s", genRandNumStr(27)),
		SAMLAssertionID:       fmt.Sprintf("id%s", genRandNumStr(27)),
		IssueInstant:          testTimeNow.Format(time.RFC3339),
		OKTAID:                fmt.Sprintf("http://www.okta.com/%s", genRandStr(20)),
		NameID:                genRandStr(8),
		ConditionNotBefore:    testTimeNow.Add(time.Microsecond * 2).Format(time.RFC3339),
		ConditionNotOnOrAfter: testTimeNow.Add(time.Minute * 10).Format(time.RFC3339),
		SessionIndex:          fmt.Sprintf("id%s.%s", genRandNumStr(13), genRandNumStr(10)),
		AWSRole:               fmt.Sprintf("arn:aws:iam::%s:saml-provider/okta,arn:aws:iam::%s:role/okta-admin-role", awsID, awsID),
	}

	tplSAMLValue := template.Must(template.New("saml-value").Parse(testSAMLResponseValueXML))
	var valueBuf bytes.Buffer
	err := tplSAMLValue.Execute(&valueBuf, tps)
	assert.NoError(t, err, "not expecting an error here")

	valueStr := base64.StdEncoding.EncodeToString(valueBuf.Bytes())
	valueStr = strings.Replace(valueStr, "+", "&#x2b;", -1)
	valueStr = strings.Replace(valueStr, "+", "&#x3d;", -1)

	tph := tplDataForSAMLHTML{
		SAMLValue: valueStr,
	}
	tplSAMLHTML := template.Must(template.New("saml-html").Parse(testSAMLHTML))
	var htmlBuf bytes.Buffer
	err = tplSAMLHTML.Execute(&htmlBuf, tph)
	assert.NoError(t, err, "not expecting an error here")

	got := SAMLAssertion{}
	err = ParseSAML(htmlBuf.Bytes(), &got)
	assert.NoError(t, err, "not expecting an error here")

	fmt.Printf("got.Resp: %#v\n", got.Resp)
	assert.Equal(t, got.Resp.SAMLP, "", "should be empty")
	assert.Equal(t, got.Resp.SAML, "", "should be empty")
	assert.Equal(t, got.Resp.SAMLSIG, "", "should be empty")
	assert.Equal(t, got.Resp.Destination, tps.Dest, "should be the same")
	assert.Equal(t, got.Resp.ID, tps.SAMLResponseID, "should be the same")
	assert.Equal(t, got.Resp.Version, tps.Version, "should be the same")
	assert.Equal(t, got.Resp.IssueInstant, tps.IssueInstant, "should be the same")
	assert.Equal(t, got.Resp.InResponseTo, "", "should be empty")
}

type testCaseGR struct {
	name, role, principal string
	shouldError           bool
}

func TestGetRole(t *testing.T) {
	awsID := genRandNumStr(12)
	tar := saml.AssumableRoles{
		saml.AssumableRole{
			Role:      fmt.Sprintf("arn:aws:iam::%s:role/okta-viewonly-role", awsID),
			Principal: fmt.Sprintf("arn:aws:iam::%s:saml-provider/okta", awsID),
		},
		saml.AssumableRole{
			Role:      fmt.Sprintf("arn:aws:iam::%s:role/okta-admin-role", awsID),
			Principal: fmt.Sprintf("arn:aws:iam::%s:saml-provider/okta", awsID),
		},
	}
	testCases := []testCaseGR{
		testCaseGR{
			name:        "find role/principal in list",
			role:        fmt.Sprintf("arn:aws:iam::%s:role/okta-viewonly-role", awsID),
			principal:   fmt.Sprintf("arn:aws:iam::%s:saml-provider/okta", awsID),
			shouldError: false,
		},
		testCaseGR{
			name:        "handle role/principal not in list",
			role:        fmt.Sprintf("arn:aws:iam::%s:role/read-only-role", awsID),
			principal:   fmt.Sprintf("arn:aws:iam::%s:saml-provider/acme", awsID),
			shouldError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetRole(tar, tc.role)
			if err != nil && !tc.shouldError {
				t.Fatal(err)
			}
			if err == nil && tc.shouldError {
				t.Errorf("was expecting to error but did not, instead got: %#v", got)
			}
			if !tc.shouldError {
				assert.Equal(t, got.Role, tc.role, "should be the same")
				assert.Equal(t, got.Principal, tc.principal, "should be the same")
			} else {
				assert.Equal(t, got.Role, "", "should be empty")
				assert.Equal(t, got.Principal, "", "should be empty")
			}
		})
	}
	t.Run("empty role list", func(t *testing.T) {
		got, err := GetRole(saml.AssumableRoles{}, "")
		assert.Error(t, err, "expecting error")
		assert.Equal(t, got.Role, "", "should be empty")
		assert.Equal(t, got.Principal, "", "should be empty")
	})
	t.Run("single role", func(t *testing.T) {
		ar := saml.AssumableRole{
			Role:      fmt.Sprintf("arn:aws:iam::%s:role/okta-admin-role", awsID),
			Principal: fmt.Sprintf("arn:aws:iam::%s:saml-provider/okta", awsID),
		}
		got, err := GetRole(saml.AssumableRoles{ar}, "")
		assert.NoError(t, err)
		assert.Equal(t, got.Role, ar.Role, "should be the same")
		assert.Equal(t, got.Principal, ar.Principal, "should be the same")
	})
}

type testCaseAIDR struct {
	name        string
	given       string
	shouldError bool
}

func TestAccountIDAndRoleFromRoleARN(t *testing.T) {
	accountID := genRandNumStr(12)
	roleName := genRandStr(5)

	testCases := []testCaseAIDR{
		testCaseAIDR{
			name: "format resource-id",
			given: fmt.Sprintf(
				"arn:partition:iam::%s:%s",
				accountID,
				roleName),
			shouldError: true,
		},
		testCaseAIDR{
			name: "format resource-type/resource-id",
			given: fmt.Sprintf(
				"arn:partition:iam::%s:role/%s",
				accountID,
				roleName),
			shouldError: false,
		},
		testCaseAIDR{
			name: "format resource-type:resource-id",
			given: fmt.Sprintf(
				"arn:partition:iam::%s:role:%s",
				accountID,
				roleName),
			shouldError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotAccountID, gotRoleName := accountIDAndRoleFromRoleARN(tc.given)
			if tc.shouldError {
				assert.Equal(t, "", gotAccountID, "should be empty")
				assert.Equal(t, tc.given, gotRoleName, "should be the same")
			} else {
				assert.Equal(t, accountID, gotAccountID, "should be the same")
				assert.Equal(t, roleName, gotRoleName, "should be the same")
			}
		})
	}
}

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const numbs = "01233456789"

func genRandStr(n int) string {
	var buf bytes.Buffer
	charsAndNumbs := numbs + chars
	for i := 0; i < n; i++ {
		buf.WriteString(string(charsAndNumbs[rand.Intn(len(charsAndNumbs))]))
	}
	return buf.String()
}

func genRandNumStr(n int) string {
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		buf.WriteString(string(numbs[rand.Intn(len(numbs))]))
	}
	return buf.String()
}
