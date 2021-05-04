package main

import (
    "encoding/json"
    "log"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
    "os"

    "github.com/olekukonko/tablewriter"
    "github.com/manifoldco/promptui"
)

type StatesResponse struct {
    States []struct {
        StateID   int    `json:"state_id"`
        StateName string `json:"state_name"`
    } `json:"states"`
    TTL int `json:"ttl"`
}

type DistrictsResponse struct {
    Districts []struct {
        DistrictID   int    `json:"district_id"`
        DistrictName string `json:"district_name"`
    } `json:"districts"`
    TTL int `json:"ttl"`
}

type SlotsResponse struct {
    Centers []struct {
        CenterID     int    `json:"center_id"`
        Name         string `json:"name"`
        StateName    string `json:"state_name"`
        DistrictName string `json:"district_name"`
        BlockName    string `json:"block_name"`
        Pincode      json.Number `json:"pincode"`
        Lat          int    `json:"lat"`
        Long         int    `json:"long"`
        From         string `json:"from"`
        To           string `json:"to"`
        FeeType      string `json:"fee_type"`
        Sessions     []struct {
            SessionID         string   `json:"session_id"`
            Date              string   `json:"date"`
            AvailableCapacity json.Number   `json:"available_capacity,omitempty"`
            MinAgeLimit       json.Number   `json:"min_age_limit"`
            Vaccine           string   `json:"vaccine,omitempty"`
            Slots             []string `json:"slots,omitempty"`
        } `json:"sessions"`
    } `json:"centers"`
}

func Request(reqUrl string) ( []byte, error ) {
    client := &http.Client{}
    req, err := http.NewRequest("GET", reqUrl, nil)
    if err != nil {
        return []byte{}, err
    }
    resp, err := client.Do(req)
    if err != nil {
        return []byte{}, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return []byte{}, err
    }

    return body, nil
}

func GetStatesList() (map[string]int, error) {
    reqUrl := "https://cdn-api.co-vin.in/api/v2/admin/location/states"

    body, err := Request(reqUrl)
    if err != nil {
        return nil, err
    }

    var r StatesResponse
    if err := json.Unmarshal(body, &r); err != nil {
        return nil, err
    }
    s := make(map[string]int)
    for _, v := range r.States {
        s[v.StateName] = v.StateID
    }
    return s, nil
}

func GetDistricts( stateId int) (map[string]int, error) {
    reqUrl := fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/admin/location/districts/%v", stateId)
    body, err := Request(reqUrl)
    if err != nil {
        return nil, err
    }
    
    var r DistrictsResponse
    if err := json.Unmarshal(body, &r); err != nil {
        return nil, err
    }
    s := make(map[string]int)
    for _, v := range r.Districts {
        s[v.DistrictName] = v.DistrictID
    } 
    return s, nil
}

func GetSlots( districId int) ( *SlotsResponse, error ) {
    today := time.Now().Format("02-01-2006")
    // log.Printf("%s - %s", today, districId)
    
    reqUrl := fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/appointment/sessions/calendarByDistrict?district_id=%v&date=%s", districId, today)        
    body, err := Request(reqUrl)
    if err != nil {
        return nil, err
    }

    /* jsonFile, _ := os.Open("/home/prajith/x")
    body, _ := ioutil.ReadAll(jsonFile) */

    var s *SlotsResponse
    if err := json.Unmarshal(body, &s); err != nil {
        return nil, err
    }

    return s, nil
}

func prompt(label string, data map[string]int) (string, error) {
    items := make([]string, 0, len(data))
    for key := range data {
        items = append(items, key)
    }

    prompt := promptui.Select{
        Label: label,
        Items: items,
    }
    _, result, err := prompt.Run()
    if err != nil {
        return "", err 
    }

    return result, nil
}

func main() {
    states, err := GetStatesList()
    if err != nil {
       log.Fatalf("Get States List Error: %s", err)
    }

    s, err := prompt("Select States", states)
    if err != nil {
        log.Fatalf("Prompt failed %v", err)
    } 
    stateId := states[s]

    districts, _ := GetDistricts(stateId)
    if err != nil {
       log.Fatalf("Get States List Error: %s", err)
    }
    
    r, err := prompt("Select Districts", districts)
    if err != nil {
        log.Fatalf("Prompt failed %v", err)
    }
    districtId := districts[r]

    slots, err := GetSlots(districtId)
    if err != nil {
       log.Fatalf("There's an error when listing slots")
    }

    if len(slots.Centers) < 1 {
        log.Fatalf("There are no vaccinations centers are available")
    }

    data := [][]string{}
    for _, center := range slots.Centers {
       for _, session := range center.Sessions {
           data = append(data, []string{ 
                session.Date, 
                center.Name, 
                center.BlockName, 
                string(center.Pincode), 
                session.Vaccine, 
                string(session.MinAgeLimit), 
                string(session.AvailableCapacity),
           })
       }
    }

    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Date", "Name", "Block", "Pincode", "Vaccine", "AgeLimit", "AvailableCapacity"})
    table.SetAutoMergeCellsByColumnIndex([]int{1})
    table.SetRowLine(true)
    table.AppendBulk(data)
    table.Render()
}
