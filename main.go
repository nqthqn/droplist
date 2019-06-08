package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/digitalocean/godo"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
)

type configuration struct {
	PersonalAccessToken string
}
type tokenSource struct {
	AccessToken string
}

func main() {
	systray.Run(renderList, func() { fmt.Println("Goodbye.") })
}

func renderList() {
	PAT := getTokenFromFile()
	client := authenticateClient(PAT)
	dropletList, _ := dropletList(client)

	systray.SetTitle("💧")
	systray.SetTooltip("You have " + strconv.Itoa(len(dropletList)) + " droplets")
	mItem, dropletURL := getDropletMenuItem(dropletList[0])

	for {
		select {
		case <-mItem.ClickedCh:
			open.Run(dropletURL)
		}
	}
}

func getDropletMenuItem(droplet godo.Droplet) (item *systray.MenuItem, url string) {
	name := droplet.Name
	ip := droplet.Networks.V4[0].IPAddress
	region := getFlagByRegionSlug(droplet.Region.Slug)
	itemText := fmt.Sprintf("%s - %s %s", name, ip, region)

	item = systray.AddMenuItem(itemText, "Quit it!")
	url = "https://cloud.digitalocean.com/droplets/" +
		strconv.Itoa(droplet.ID)

	return
}

func getFlagByRegionSlug(region string) string {
	flags := map[string]string{
		"DE": "\U0001F1E9\U0001F1EA",
		"US": "\U0001F1FA\U0001F1F8",
		"SG": "\U0001F1F8\U0001F1EC",
		"GB": "\U0001F1EC\U0001F1E7",
		"NL": "\U0001F1F3\U0001F1F1",
	}
	regionsToCountries := map[string]string{
		"fra": "DE",
		"nyc": "US",
		"sfo": "US",
		"ams": "NL",
		"sgp": "SG",
		"lon": "GB",
	}

	simpleRegion := region[:3]
	country := regionsToCountries[simpleRegion]

	return flags[country]
}

func dropletList(client *godo.Client) ([]godo.Droplet, error) {
	list := []godo.Droplet{}
	opt := &godo.ListOptions{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		droplets, resp, err := client.Droplets.List(ctx, opt)
		if err != nil {
			return nil, err
		}

		for _, d := range droplets {
			list = append(list, d)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return list, nil
}

/*
* Authenticating client to use DO API
 */
func getTokenFromFile() string {
	// Get PersonalAccessToken from config file
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Println("could not open config.json:", err)
		os.Exit(1)
	}
	decoder := json.NewDecoder(file)
	config := configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("could not parse config.json:", err)
		os.Exit(1)
	}

	return config.PersonalAccessToken
}

func authenticateClient(accessToken string) (client *godo.Client) {
	tokenSource := &tokenSource{
		AccessToken: accessToken,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client = godo.NewClient(oauthClient)
	return
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}
