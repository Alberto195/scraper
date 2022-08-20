package scrapers

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"hello/scraper/models"
	scrapers "hello/scraper/scrapers/mock"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

//go:generate mockgen -source=parser.go -destination=./mock/parser_httpclient_mock.go -package=scrapers

func TestParser_NewOIDParser(t *testing.T) {
	parserExpected := &OidParser{
		urlCache:   nil,
		httpClient: http.DefaultClient,
	}
	parserActual := NewOIDParser(nil, http.DefaultClient)

	assert.Equal(t, parserExpected, parserActual)
}

func TestParser_filter(t *testing.T) {
	parser := NewOIDParser(&sync.Map{}, http.DefaultClient)
	tests := []struct {
		name         string
		body         string
		expectedData map[string]*models.TableInfo
	}{
		{
			name: "success",
			body: "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <title> Global OID reference database </title>\n    <meta name=\"description\" content=\"\">\n\n    <script src=\"/cdn-cgi/apps/head/2VsPAxpuBO-CkkZXqaeHnqT5qxU.js\"></script><script>\n      (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){\n      (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),\n      m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)\n      })(window,document,'script','https://www.google-analytics.com/analytics.js','ga');\n\n      ga('create', 'UA-82642346-1', 'auto');\n      ga('send', 'pageview');\n    </script>\n\n    \n    <meta name=\"google-site-verification\" content=\"goC5jUiwFWTihZyBplddmH71LTkzQSVB89OWNoZKbEU\" />\n    <meta name=\"yandex-verification\" content=\"0c17a632986db6ed\" />\n\n\n    <style>\n        \n        \n        table {\n            border: solid brown 1px;\n            border-collapse: collapse;\n            margin-top: 15px\n        }\n        table td {\n            border: solid brown 1px;\n            padding: 2px 3px;\n        }\n        table th {\n            border: solid brown 1px;\n            padding: 5px 5px;\n            color: white;\n            background-color: hsla(34,85%,45%,1);\n            font-weight: normal;\n        }\n        h1 {\n            text-align: left;\n            display: block;\n            width: 100%;\n        }\n        h3 {\n            text-align: left;\n            display: block;\n            width: 100%;\n            background-color: #58bdff;\n            padding: 5px 0 5px 15px;\n            margin: 15px 0 0 0;\n        }\n        p {\n            padding: 5px 0 5px 15px;\n            margin: 0;\n            border-left: solid 2px #58bdff;\n        }\n        dl {\n          width: 100%;\n          overflow: hidden;\n          //background: #ff0;\n          padding: 0;\n          margin-top: 15px;\n        }\n        dt {\n          //float: left;\n          width: 20%;\n          background: #ff9400;\n          border-left: solid 2px #9b5700;\n          font-weight: bolder;\n          text-align: center;\n          padding: 5px 10px 5px 0;\n          margin-top: 15px;\n        }\n        dd {\n          //float: left;\n          width: 75%;\n          //background: #dd0;\n          border-left: solid 2px #9b5700;\n          padding: 5px 0 0 20px;\n          margin: 0;\n        }\n        .breadcrumb {\n            list-style: none;\n            overflow: hidden;\n            font: 16px Helvetica, Arial, Sans-Serif;\n            margin: 0;\n            padding: 0;\n        }\n        .breadcrumb li {\n            float: left;\n        }\n        .breadcrumb li a {\n            color: white;\n            text-decoration: none;\n            padding: 5px 0 5px 45px;\n            background: brown;                   /* fallback color */\n            background: hsla(34,85%,35%,1);\n            position: relative;\n            display: block;\n            float: left;\n        }\n\n        .breadcrumb li a:after {\n            content: \" \";\n            display: block;\n            width: 0;\n            height: 0;\n            border-top: 50px solid transparent;           /* Go big on the size, and let overflow hide */\n            border-bottom: 50px solid transparent;\n            border-left: 30px solid hsla(34,85%,35%,1);\n            position: absolute;\n            top: 50%;\n            margin-top: -50px;\n            left: 100%;\n            z-index: 2;\n        }\n\n        .breadcrumb li a:before {\n            content: \" \";\n            display: block;\n            width: 0;\n            height: 0;\n            border-top: 50px solid transparent;\n            border-bottom: 50px solid transparent;\n            border-left: 30px solid white;\n            position: absolute;\n            top: 50%;\n            margin-top: -50px;\n            margin-left: 1px;\n            left: 100%;\n            z-index: 1;\n        }\n\n        .breadcrumb li:first-child a {\n            padding-left: 10px;\n        }\n        .breadcrumb li:nth-child(2) a       { background:        hsla(34,85%,45%,1); }\n        .breadcrumb li:nth-child(2) a:after { border-left-color: hsla(34,85%,45%,1); }\n        .breadcrumb li:nth-child(3) a       { background:        hsla(34,85%,55%,1); }\n        .breadcrumb li:nth-child(3) a:after { border-left-color: hsla(34,85%,55%,1); }\n        .breadcrumb li:nth-child(4) a       { background:        hsla(34,85%,65%,1); }\n        .breadcrumb li:nth-child(4) a:after { border-left-color: hsla(34,85%,65%,1); }\n        .breadcrumb li:nth-child(5) a       { background:        hsla(34,85%,67%,1); }\n        .breadcrumb li:nth-child(5) a:after { border-left-color: hsla(34,85%,67%,1); }\n        .breadcrumb li:nth-child(6) a       { background:        hsla(34,85%,69%,1); }\n        .breadcrumb li:nth-child(6) a:after { border-left-color: hsla(34,85%,69%,1); }\n        .breadcrumb li:nth-child(7) a       { background:        hsla(34,85%,72%,1); }\n        .breadcrumb li:nth-child(7) a:after { border-left-color: hsla(34,85%,72%,1); }\n        .breadcrumb li:nth-child(8) a       { background:        hsla(34,85%,74%,1); }\n        .breadcrumb li:nth-child(8) a:after { border-left-color: hsla(34,85%,74%,1); }\n        .breadcrumb li:last-child a {\n            #background: transparent !important;\n            #color: black;\n            pointer-events: none;\n            cursor: default;\n        }\n\n        .breadcrumb li a:hover { background: hsla(34,85%,25%,1); }\n        .breadcrumb li a:hover:after { border-left-color: hsla(34,85%,25%,1) !important; }\n\n        /* CSSTerm.com Simple CSS menu */\n\n        #br { clear:left }\n\n        .menu_simple {\n            width: 100%;\n            background-color: #005555;\n        }\n\n        .menu_simple ul {\n            margin: 0; padding: 0;\n            float: left;\n        }\n\n        .menu_simple ul li {\n            display: inline;\n        }\n\n        .menu_simple ul li a {\n            float: left; text-decoration: none;\n            color: white;\n            padding: 10.5px 11px;\n            background-color: #417690;\n        }\n\n        .menu_simple ul li a:visited {\n            color: white;\n        }\n\n        .menu_simple ul li a:hover, .menu_simple ul li .current {\n            color: white;\n            background-color: #5FD367;\n        }\n    </style>\n</head>\n<body>\n\n    <div class=\"menu_simple\">\n    <ul>\n        <li><a href=\"/\">Main page</a></li>\n        <li><a href=\"/orgs/\">Organizations list</a></li>\n        <li><a href=\"/contacts\">Contacts</a></li>\n    </ul>\n    </div>\n    <div style=\"clear: both\"></div>\n\n\n<h1>Global OID reference database</h1>\n\n<p>This is full world OID database published for internet users</p>\n\n<h2>Root Tree Nodes</h2>\n<table>\n    <tr><th>Node</th><th>Name</th><th>Sub children</th><th>Sub Nodes Total</th><th>Description</th><th>Information</th></tr>\n    <tr><td><a href=\"/0\">0</a></td><td>itu-t, ccitt</td><td>7</td><td>10360</td><td>International Telecommunications Union - Telecommunication standardization sector (ITU-T)</td><td>Subsequent OIDs identify ITU-T Recommendations (not jointly published with ISO/IEC) and ITU members.<br>\n<br>\nThis arc is also called <code>ccitt(0)</code> to recall that CCITT used to be an organization independent from ITU-T.<br>\n<br>\nIdentifier <strong><code>itu-r</code></strong> was added by ITU-T Study Group 17 in March 2004 (and was ratified by ISO/IEC JTC 1/SC 6 in Sep 2005). It can only be used as a 'NameAndNumberForm' (that is, followed by number <code>5</code> between parentheses) for OIDs that commence with <code>{itu-r(0) <a href=\"https://oidref.com/0.5\">r-recommendation(5)</a>}</code> (see <a href=\"http://itu.int/rec/T-REC-X.680/en\">Rec. ITU-T X.680 | ISO/IEC 9834-1</a>, clause A.5, for more details on this specific case). Consequently Unicode label <code>ITU-R</code> can only be used for \"<a href=\"http://oid-info.com/faq.htm#iri\">OID-IRIs</a>\" that designate OIDs under the <code>{itu-r(0) <a href=\"https://oidref.com/0.5\">r-recommendation(5)</a>}</code> arc.<br>\n<br>\nOperation is in accordance with <a href=\"http://itu.int/rec/T-REC-X.660/en\">Rec. ITU-T X.660 | ISO/IEC 9834-1</a> and is under the guidance of <a href=\"http://itu.int/ITU-T/studygroups/com17/index.asp\">ITU-T Study Group 17</a>.<br>\n<br>\nAll decisions related to subsequent arcs, other than the assignment of additional secondary identifiers to top-level arc <code>0</code> (see Rec. ITU-T X.660 | ISO/IEC 9834-1, clause A.5), will be recorded ad amendments to Rec. ITU-T X.660 | ISO/IEC 9834-1 (such changes to the joint ITU-T | ISO/IEC text will be regarded as editorial by ISO).<br>\n<br>\nFrom Rec. ITU-T X.660 | ISO/IEC 9834-1, \"the top-level arcs are restricted to three arcs numbered <code>0</code> to <code>2</code>; and the arcs beneath root arcs <code>0</code> and <code>1</code> are restricted to forty arcs numbered <code>0</code> to <code>39</code>. This enables optimized encodings to be used in which the values of the top two arcs for all arcs under top-level arcs <code>0</code> and <code>1</code> encode in a single octet in an object identifier encoding (see the Rec. ITU-T X.690 series | ISO/IEC 8825 multi-part Standard).</td></tr>\n    <tr><td><a href=\"/1\">1</a></td><td>iso</td><td>4</td><td>992195</td><td>International Organization for Standardization (ISO)</td><td>This arc is for International Standards and ISO Member Bodies.<br>\n<br>\nOperation of this arc is in accordance with <a href=\"http://itu.int/ITU-T/X.660\">Rec. ITU-T X.660 | ISO/IEC 9834-1</a> \"<em>Procedures for the operation of object identifier registration authorities: General procedures and top arcs of the international object identifier tree</em>\".<br>\n<br>\nAll decisions related to subsequent arcs, other than the assignment of additional secondary identifiers to top-level arc <code>1</code> (see Rec. ITU-T X.660 (2004) | ISO/IEC 9834-1:2004, A.5), will be recorded as amendments to Rec. ITU-T X.660 | ISO/IEC 9834-1 (such changes to the common text will be regarded as editorial by ITU-T).<br>\n<br>\nFrom Rec. ITU-T X.660 (2004) | ISO/IEC 9834-1:2004, \"the top-level arcs are restricted to three arcs numbered 0 to 2; and the arcs beneath root arcs <code>0</code> and <code>1</code> are restricted to forty arcs numbered <code>0</code> to <code>39</code>. This enables optimized encodings to be used in which the values of the top two arcs for all arcs under top-level arcs <code>0</code> and <code>1</code> encode in a single octet in an object identifier encoding (see the Rec. ITU-T X.690 series | ISO/IEC 8825 multi-part Standard).</td></tr>\n    <tr><td><a href=\"/2\">2</a></td><td>joint-iso-itu-t, joint-iso-ccitt</td><td>38</td><td>25835</td><td>Common standardization area of ISO/IEC (International Organization for Standardization/International Electrotechnical Commission) and ITU-T (International Telecommunications Union - Telecommunication standardization sector)</td><td>This OID was allocated by <a href=\"http://itu.int/ITU-T/X.660\">Rec. ITU-T X.660</a> | ISO/IEC 9834-1.<br>\n<br>\nThis OID is jointly administered by ISO and ITU-T according to <a href=\"http://itu.int/ITU-T/X.662\">Rec. ITU-T X.662</a> | ISO/IEC 9834-3 \"<em>Procedures for the Operation of OSI Registration Authorities: Registration of Object Identifier Arcs for Joint ISO and ITU-T Work</em>\". As a consequence, all requests for registration must be jointly approved by ITU-T Study Group 17 and ISO/IEC JTC 1/SC 6. Child OIDs are recorded in the <a href=\"http://itu.int/go/X660\">Register of arcs beneath the root arc with primary integer value 2</a>.<br>\n<br>\nNew child OIDs will be allocated a number greater than 47, except if there is a good rationale that a compact binary encoding is needed, in which case a number less or equal to 47 can be allocated so that the OID encodes with a single octet.</td></tr>\n</table>\n\n\n\n\n</body>\n</html>",
			expectedData: map[string]*models.TableInfo{
				"/": {
					Name:     "",
					SubCh:    0,
					SubTotal: 0,
					Desc:     "",
					Inf:      "",
				},
				"/0": {
					Name:     "itu-t, ccitt",
					SubCh:    7,
					SubTotal: 10360,
					Desc:     "International Telecommunications Union - Telecommunication standardization sector (ITU-T)",
					Inf:      "Subsequent OIDs identify ITU-T Recommendations (not jointly published with ISO/IEC) and ITU members.",
				},
				"/1": {
					Name:     "iso",
					SubCh:    4,
					SubTotal: 992195,
					Desc:     "International Organization for Standardization (ISO)",
					Inf:      "This arc is for International Standards and ISO Member Bodies.",
				}, "/2": {
					Name:     "joint-iso-itu-t, joint-iso-ccitt",
					SubCh:    38,
					SubTotal: 25835,
					Desc:     "Common standardization area of ISO/IEC (International Organization for Standardization/International Electrotechnical Commission) and ITU-T (International Telecommunications Union - Telecommunication standardization sector)",
					Inf:      "This OID was allocated by ",
				}},
		},
		{
			name:         "bad text",
			body:         "error text",
			expectedData: make(map[string]*models.TableInfo, 10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := parser.filter([]byte(tt.body))

			assert.Equal(t, data, tt.expectedData)
			assert.NoError(t, err)
		})
	}
}

func TestParser_getBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name  string
		init  func(mockHttpClient *scrapers.MockHTTPClient)
		url   string
		error error
	}{
		{
			name: "success",
			init: func(mockHttpClient *scrapers.MockHTTPClient) {
				r := strings.NewReader("")
				mockHttpClient.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(r)}, nil)
			},
			url:   "/",
			error: nil,
		},
		{
			name: "request error",
			init: func(mockHttpClient *scrapers.MockHTTPClient) {
				mockHttpClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("test error"))
			},
			url:   "/error",
			error: errors.New("test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := scrapers.NewMockHTTPClient(ctrl)
			tt.init(mockHttpClient)
			parser := NewOIDParser(&sync.Map{}, mockHttpClient)

			_, err := parser.getBody(tt.url)
			if tt.error != nil {
				assert.EqualError(t, err, tt.error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParser_mapToTable(t *testing.T) {
	tests := []struct {
		name     string
		data     []string
		expected *models.TableInfo
	}{
		{
			name: "1",
			data: []string{"1"},
			expected: &models.TableInfo{
				Name: "1",
			},
		},
		{
			name: "2",
			data: []string{"1", "0"},
			expected: &models.TableInfo{
				Name:  "1",
				SubCh: 0,
			},
		},
		{
			name: "3",
			data: []string{"1", "0", "10"},
			expected: &models.TableInfo{
				Name:     "1",
				SubCh:    0,
				SubTotal: 10,
			},
		},
		{
			name: "4",
			data: []string{"1", "0", "10", "description"},
			expected: &models.TableInfo{
				Name:     "1",
				SubCh:    0,
				SubTotal: 10,
				Desc:     "description",
			},
		},
		{
			name: "5",
			data: []string{"1", "0", "10", "description", "inf"},
			expected: &models.TableInfo{
				Name:     "1",
				SubCh:    0,
				SubTotal: 10,
				Desc:     "description",
				Inf:      "inf",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := mapToTableInfo(tt.data)
			assert.Equal(t, tt.expected, model)
		})
	}
}
