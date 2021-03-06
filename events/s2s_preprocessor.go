package events

import (
	"errors"
	"github.com/ksensehq/eventnative/appconfig"
	"github.com/ksensehq/eventnative/geo"
	"github.com/ksensehq/eventnative/useragent"
	"log"
	"net/http"
)

//S2SPreprocessor preprocess server 2 server integration events
type S2SPreprocessor struct {
	geoResolver geo.Resolver
	uaResolver  useragent.Resolver
}

func NewS2SPreprocessor() Preprocessor {
	return &S2SPreprocessor{
		geoResolver: appconfig.Instance.GeoResolver,
		uaResolver:  appconfig.Instance.UaResolver,
	}
}

//Preprocess resolve geo from ip field or skip if geo.GeoDataKey field was provided
//resolve useragent from uaKey or skip if useragent.ParsedUaKey field was provided
//transform data to c2s format
func (s2sp *S2SPreprocessor) Preprocess(fact Fact, r *http.Request) (Fact, error) {
	if fact == nil {
		return nil, errors.New("Input fact can't be nil")
	}

	processed := Fact{}

	eventCtx := map[string]interface{}{}

	processed["event_data"] = fact["event_data"]
	processed["src"] = "s2s"
	processed[eventnKey] = eventCtx

	eventCtx["event_id"] = fact["event_id"]
	eventCtx["user"] = fact["user"]
	if pageCtx, ok := fact["page_ctx"]; ok {
		if pageCtxObject, ok := pageCtx.(map[string]interface{}); ok {
			for k, v := range pageCtxObject {
				eventCtx[k] = v
			}
		}
	}

	if deviceCtx, ok := fact["device_ctx"]; ok {
		if deviceCtxObject, ok := deviceCtx.(map[string]interface{}); ok {
			//geo.GeoDataKey node overwrite geo resolving
			if location, ok := deviceCtxObject[geo.GeoDataKey]; !ok {
				if ip, ok := deviceCtxObject["ip"]; ok {
					geoData, err := s2sp.geoResolver.Resolve(ip.(string))
					if err != nil {
						log.Println(err)
					}

					eventCtx[geo.GeoDataKey] = geoData
				}
			} else {
				eventCtx[geo.GeoDataKey] = location
			}

			//useragent.ParsedUaKey node overwrite useragent resolving
			if parsedUa, ok := deviceCtxObject[useragent.ParsedUaKey]; !ok {
				if ua, ok := deviceCtxObject[uaKey]; ok {
					eventCtx[uaKey] = ua
					if uaStr, ok := ua.(string); ok {
						eventCtx[useragent.ParsedUaKey] = s2sp.uaResolver.Resolve(uaStr)
					}
				}
			} else {
				eventCtx[useragent.ParsedUaKey] = parsedUa
			}
		}
	}

	return processed, nil
}
