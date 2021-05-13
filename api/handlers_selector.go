package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (s *server) InstanceSelectorHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := vars["account"]

	queries := r.URL.Query()
	if len(queries) == 0 {
		handleError(w, apierror.New(apierror.ErrBadRequest, "filter is required", nil))
		return
	}

	filters, err := parseFilterQueries(r.Context(), queries)
	if err != nil {
		handleError(w, err)
		return
	}

	role := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName)

	session, err := s.assumeRole(
		r.Context(),
		s.session.ExternalID,
		role,
		"",
		"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess",
	)
	if err != nil {
		msg := fmt.Sprintf("failed to assume role in account: %s", account)
		handleError(w, apierror.New(apierror.ErrForbidden, msg, err))
		return
	}

	instanceSelector := selector.New(session.Session)

	out, err := instanceSelector.Filter(filters)
	if err != nil {
		fmt.Printf("Oh no, there was an error :( %v", err)
		return
	}

	handleResponseOk(w, out)
}

// parseFilterQueries parses the filters passed as query parameters
func parseFilterQueries(ctx context.Context, q url.Values) (selector.Filters, error) {
	log.Debugf("parsing query values: %+v", q)

	var allowList, denylist *regexp.Regexp

	// build a map of filters
	filtersMap := map[string]interface{}{}
	for k, values := range q {
		log.Debugf("parsing query %s:%+v", k, values)

		// availabilityzone is the only []*string filter and is treated as a special case
		if l := strings.ToLower(k); l == "availabilityzones" {
			log.Debugf("%s: %+v is an availability zone list", k, values)
			filtersMap[k] = values
			continue
		}

		value := values[0]

		// if it's one of the filters that takes a regexp
		if l := strings.ToLower(k); l == "allowlist" || l == "denylist" {
			log.Debugf("%s: %s is a regular expression allowlist/denylist", k, value)
			if r, err := regexp.Compile(value); err == nil {
				switch l {
				case "allowlist":
					allowList = r
				case "denylist":
					denylist = r
				}
			} else {
				log.Warnf("%s is not a valid regular expression: %s", value, err)
			}

			continue
		}

		if i, err := strconv.Atoi(value); err == nil {
			log.Debugf("%s: %s is an int", k, value)
			filtersMap[k] = aws.Int(i)
			continue
		}

		if f, err := strconv.ParseFloat(value, 64); err == nil {
			log.Debugf("%s: %s is a float", k, value)
			filtersMap[k] = aws.Float64(f)
			continue
		}

		if b, err := strconv.ParseBool(value); err == nil {
			log.Debugf("%s: %s is a boolean", k, value)
			filtersMap[k] = aws.Bool(b)
			continue
		}

		if strings.Contains(value, "-") {
			ranges := strings.SplitN(value, "-", 2)

			isInt := true
			low, err := strconv.Atoi(ranges[0])
			if err != nil {
				isInt = false
			}

			high, err := strconv.Atoi(ranges[1])
			if err != nil {
				isInt = false
			}

			// if we have an int range
			if isInt {
				log.Debugf("%s: %s is an int range", k, value)
				filtersMap[k] = &selector.IntRangeFilter{
					UpperBound: high,
					LowerBound: low,
				}
				continue
			}

			lowb, err := bytequantity.ParseToByteQuantity(ranges[0])
			if err != nil {
				log.Warnf("%s: %s failed to convert to byte quantity", k, ranges[0])
				continue
			}

			highb, err := bytequantity.ParseToByteQuantity(ranges[1])
			if err != nil {
				log.Warnf("%s: %s failed to convert to byte quantity", k, ranges[1])
				continue
			}

			log.Debugf("%s: %s is a byte quantity range (%d/%d)", k, value, lowb.Quantity, highb.Quantity)

			// if we have a byte quantity range
			filtersMap[k] = &selector.ByteQuantityRangeFilter{
				UpperBound: highb,
				LowerBound: lowb,
			}

			continue
		}

		filtersMap[k] = aws.String(value)
	}

	log.Debugf("converting to json: %+v", filtersMap)

	// convert to JSON and back to filters, let the json encoder/decoder do the work
	jsonOut, err := json.Marshal(filtersMap)
	if err != nil {
		return selector.Filters{}, err
	}

	log.Debugf("json: %s", string(jsonOut))

	var filters selector.Filters
	if err := json.Unmarshal(jsonOut, &filters); err != nil {
		return selector.Filters{}, err
	}

	// can't marshal/unmarshal regexp from json
	filters.AllowList = allowList
	filters.DenyList = denylist

	log.Debugf("filters: %s", awsutil.Prettify(filters))

	return filters, nil
}
