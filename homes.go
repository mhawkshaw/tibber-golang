package tibber

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"
)

// HomesResponse response from homes
type HomesResponse struct {
	Viewer HomesViewer `json:"viewer"`
}

// HomeViewer list of homes
type HomesViewer struct {
	Homes []Home `json:"homes"`
}

type HomeResponse struct {
	Viewer HomeViewer `json:"viewer"`
}

type HomeViewer struct {
	Homes []Home `json:"homes"`
	Home  Home   `json:"home"`
}

type PreviousMeterData struct {
	Power           float64 `json:"power"`
	PowerProduction float64 `json:"powerProduction"`
}

// Home structure
type Home struct {
	ID                   string                    `json:"id"`
	AppNickname          string                    `json:"appNickname"`
	MeteringPointData    MeteringPointData         `json:"meteringPointData"`
	Features             Features                  `json:"features"`
	Address              Address                   `json:"address"`
	Size                 int                       `json:"size"`
	MainFuseSize         int                       `json:"mainFuseSize"`
	NumberOfResidents    int                       `json:"numberOfResidents"`
	PrimaryHeatingSource string                    `json:"primaryHeatingSource"`
	HasVentilationSystem bool                      `json:"hasVentilationSystem"`
	CurrentSubscription  CurrentSubscription       `json:"currentSubscription"`
	PreviousMeterData    PreviousMeterData         `json:"previousMeterData"`
	Consumption          HomeConsumptionConnection `json:"consumption"`
}

type Address struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	Address3   string `json:"address3"`
	PostalCode string `json:"postalCode"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
}

// MeteringPointData - meter number
type MeteringPointData struct {
	ConsumptionEan string `json:"consumptionEan"`
}

// Features - tibber pulse connected
type Features struct {
	RealTimeConsumptionEnabled bool `json:"realTimeConsumptionEnabled"`
}

type CurrentSubscription struct {
	PriceInfo   PriceInfo   `json:"priceInfo"`
	PriceRating PriceRating `json:"priceRating"`
}

type PriceInfo struct {
	Current  Price   `json:"current"`
	Today    []Price `json:"today"`
	Tomorrow []Price `json:"tomorrow"`
}

type PriceRating struct {
	ThresholdPercentages PriceRatingThresholdPercentages `json:"thresholdPercentages"`
	Hourly               PriceRatingType                 `json:"hourly"`
	Daily                PriceRatingType                 `json:"daily"`
	Monthly              PriceRatingType                 `json:"monthly"`
}

type PriceRatingThresholdPercentages struct {
	High float64 `json:"high"`
	Low  float64 `json:"low"`
}

type PriceRatingType struct {
	MinEnergy float64            `json:"minEnergy"`
	MaxEnergy float64            `json:"maxEnergy"`
	MinTotal  float64            `json:"minTotal"`
	MaxTotal  float64            `json:"maxTotal"`
	Currency  string             `json:"currency"`
	Entries   []PriceRatingEntry `json:"entries"`
}

type PriceRatingEntry struct {
	Time       time.Time `json:"time"`
	Energy     float64   `json:"energy"`
	Total      float64   `json:"total"`
	Tax        float64   `json:"tax"`
	Difference float64   `json:"difference"`
	Level      string    `json:"level"`
}

type Price struct {
	Level    string    `json:"level"`
	Total    float64   `json:"total"`
	Energy   float64   `json:"energy"`
	Tax      float64   `json:"tax"`
	Currency string    `json:"currency"`
	StartsAt time.Time `json:"startsAt"`
}

type Consumption struct {
	From            time.Time `json:"from"`
	To              time.Time `json:"to"`
	Cost            float64   `json:"cost"`
	UnitPrice       float64   `json:"unitPrice"`
	UnitPriceVAT    float64   `json:"unitPriceVAT"`
	Consumption     float64   `json:"consumption"`
	ConsumptionUnit string    `json:"consumptionUnit"`
	Currency        string    `json:"currency"`
}

type HomeConsumptionConnection struct {
	Nodes []Consumption `json:"nodes"`
}

type PriceResolution int64

const (
	HOURLY PriceResolution = iota
	DAILY
)

func (s PriceResolution) String() string {
	switch s {
	case HOURLY:
		return "HOURLY"
	case DAILY:
		return "DAILY"
	}
	return "unknown"
}

// GetHomes get a list of homes with information
func (t *Client) GetHomes() ([]Home, error) {
	req := graphql.NewRequest(`
		query {
			viewer {
				homes {
					id
					appNickname
      				meteringPointData{
        				consumptionEan
      				}
					features {
						realTimeConsumptionEnabled
					}
					address {
						address1
						address2
						address3
						postalCode
						city
						country
						latitude
						longitude
					}
					size
					mainFuseSize
					numberOfResidents
					primaryHeatingSource
					hasVentilationSystem
				}
			}
		}`)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+t.Token)
	ctx := context.Background()
	var result HomesResponse
	if err := t.gqlClient.Run(ctx, req, &result); err != nil {
		log.Error(err)
		return nil, err
	}
	return result.Viewer.Homes, nil
}

// GetHomeById get a home with information
func (t *Client) GetHomeById(homeId string) (Home, error) {
	req := graphql.NewRequest(fmt.Sprintf(`
		query {
			viewer {
				home(id:"%s") {
					id
					appNickname
      				meteringPointData{
        				consumptionEan
      				}
					features {
						realTimeConsumptionEnabled
					}
					address {
						address1
						address2
						address3
						postalCode
						city
						country
						latitude
						longitude
					}
					size
					mainFuseSize
					numberOfResidents
					primaryHeatingSource
					hasVentilationSystem
				}
			}
		}`, homeId))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+t.Token)
	ctx := context.Background()
	var result HomeResponse
	if err := t.gqlClient.Run(ctx, req, &result); err != nil {
		log.Error(err)
		return Home{}, err
	}
	return result.Viewer.Home, nil
}

func (t *Client) GetCurrentPrice(homeId string) (Price, error) {
	req := graphql.NewRequest(fmt.Sprintf(`
		query {
			viewer {
				home(id:"%s")  {
					currentSubscription {
						priceInfo {
							current {
								level
								total
								energy
								tax
								currency
								startsAt
							}
						}
					}
				}
			}
		}`, homeId))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+t.Token)
	ctx := context.Background()
	var result HomeResponse
	if err := t.gqlClient.Run(ctx, req, &result); err != nil {
		log.Error(err)
		return Price{}, err
	}

	return result.Viewer.Home.CurrentSubscription.PriceInfo.Current, nil
}

func (t *Client) GetCurrentPriceRating(homeId string) (PriceRating, error) {
	req := graphql.NewRequest(fmt.Sprintf(`
		query {
			viewer {
				home(id:"%s")  {
					currentSubscription {
						priceRating {
							thresholdPercentages {
								high
								low
							}
							hourly {
								minEnergy
								maxEnergy
								minTotal
								currency
								entries {
									time
									energy
									total
									tax
									difference
									level
								}
							}
							daily {
								minEnergy
								maxEnergy
								minTotal
								currency
								entries {
									time
									energy
									total
									tax
									difference
									level
								}
							}
							monthly {
								minEnergy
								maxEnergy
								minTotal
								currency
								entries {
									time
									energy
									total
									tax
									difference
									level
								}
							}
						}
					}
				}
			}
		}`, homeId))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+t.Token)
	ctx := context.Background()
	var result HomeResponse
	if err := t.gqlClient.Run(ctx, req, &result); err != nil {
		log.Error(err)
		return PriceRating{}, err
	}

	return result.Viewer.Home.CurrentSubscription.PriceRating, nil
}

func (t *Client) GetTomorrowsPrice(homeId string) ([]Price, error) {
	req := graphql.NewRequest(fmt.Sprintf(`
		query {
			viewer {
				home(id:"%s")  {
					currentSubscription {
						priceInfo {
							tomorrow {
								level
								total
								energy
								tax
								currency
								startsAt
							}
						}
					}
				}
			}
		}`, homeId))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+t.Token)
	ctx := context.Background()
	var result HomeResponse
	if err := t.gqlClient.Run(ctx, req, &result); err != nil {
		log.Error(err)
		return []Price{}, err
	}

	return result.Viewer.Home.CurrentSubscription.PriceInfo.Tomorrow, nil
}

func (t *Client) GetPriceInfo(homeId string) (PriceInfo, error) {
	req := graphql.NewRequest(fmt.Sprintf(`
		query {
			viewer {
				home(id:"%s")  {
					currentSubscription {
						priceInfo {
							current {
								level
								total
								energy
								tax
								currency
								startsAt
							}
							today {
								level
								total
								energy
								tax
								currency
								startsAt
							}
							tomorrow {
								level
								total
								energy
								tax
								currency
								startsAt
							}
						}
					}
				}
			}
		}`, homeId))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+t.Token)
	ctx := context.Background()
	var result HomeResponse
	if err := t.gqlClient.Run(ctx, req, &result); err != nil {
		log.Error(err)
		return PriceInfo{}, err
	}

	return result.Viewer.Home.CurrentSubscription.PriceInfo, nil
}

func (t *Client) GetConsumption(homeId string, resolution PriceResolution, last int) ([]Consumption, error) {
	req := graphql.NewRequest(fmt.Sprintf(`
		query {
			viewer {
				home(id:"%s")  {
					consumption(resolution: %s, last: %s) {
						nodes {
							from
							to
							cost
							unitPrice
							unitPriceVAT
							currency
							consumption
							consumptionUnit
						}
					}
				}
			}
		}`, homeId, resolution, strconv.Itoa(last)))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+t.Token)
	ctx := context.Background()
	var result HomeResponse
	if err := t.gqlClient.Run(ctx, req, &result); err != nil {
		log.Error(err)
		return []Consumption{}, err
	}

	return result.Viewer.Home.Consumption.Nodes, nil
}
