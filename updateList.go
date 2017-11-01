package main

import (
    "fmt"
    "net/url"
    "sort"
    "github.com/BurntSushi/toml"
    "github.com/ChimeraCoder/anaconda"
)

type Config struct {
    ConsumerKey string
    ConsumerSecret string
    AccessToken string
    AccessTokenSecret string
    OwnerScreenName string
    ListName string
    ListId int64
    MaxCount int
    BlackList []string
}

var config = &Config{}
var api = &anaconda.TwitterApi{}

func main() {
    _, err := toml.DecodeFile("./config.tml", &config)
    if err != nil {
        panic(err)
    }
//    fmt.Printf("%+v\n", config)

    anaconda.SetConsumerKey(config.ConsumerKey)
    anaconda.SetConsumerSecret(config.ConsumerSecret)
    api = anaconda.NewTwitterApi(config.AccessToken, config.AccessTokenSecret)

    followers := getUsers("followers")
    listUsers := getUsers("listUsers")

    followerCount := len(followers)
    listUserCount := len(listUsers)
    f , lu, add, remove := 0, 0, 0, 0
    v := url.Values{}
    v.Set("owner_screen_name", config.OwnerScreenName)

    listId := int64(config.ListId)
    follower := anaconda.User{}
    listUser := anaconda.User{}

    for  {
        if f < followerCount {
            follower = followers[f]
        }
        if lu < listUserCount {
            listUser = listUsers[lu]
        }
        if f >= followerCount && lu >= listUserCount {
            break
        }
        var err interface{}
        switch {
        case follower.Id < listUser.Id || lu == listUserCount:
            _, err = api.AddListUser(follower.Id, listId, v)
            fmt.Printf("add %s %s\n", follower.ScreenName, follower.Name)
            f++
            add++
        case follower.Id > listUser.Id || f == followerCount:
            _, err = api.RemoveListUser(listUser.Id, listId, v)
            fmt.Printf("remove %s %s\n", listUser.ScreenName, listUser.Name)
            lu++
            remove++
        case follower.Id == listUser.Id:
            f++
            lu++
        }

        if err != nil {
            panic(err)
        }

    }
    fmt.Printf("follower(no protected): %d\n", followerCount)
    fmt.Printf("add: %d, remove: %d\n", add, remove)
    fmt.Printf("%s list updated.\n", config.ListName)

}

func getUsers(target string) []anaconda.User {

    v := url.Values{}
    cursor := anaconda.UserCursor{}
    cursor.Next_cursor = -1
    var err interface{}

    v.Set("screen_name", config.OwnerScreenName)
    if target == "listUsers" {
        v.Set("list_id", fmt.Sprint(config.ListId))
    }

    users := []anaconda.User{}
    for {
        if cursor.Next_cursor != -1 {
            v.Set("cursor", cursor.Next_cursor_str)
        }
        v.Set("count", fmt.Sprint(config.MaxCount))

        switch target {
        case "followers" :
            cursor, err = api.GetFollowersList(v)
        case "listUsers" :
            cursor, err = api.GetListUsers(v)
        }
        if err != nil {
            panic(err)
        }

        for _, user := range cursor.Users {
            if target == "followers" {
                skip := false
                for _, name := range config.BlackList {
                    if user.ScreenName == name {
                        skip = true
                        break
                    }
                }
                if skip { continue }
                if user.Protected == true {
                    continue
                }
            }
            users = append(users, user)
        }
        if cursor.Next_cursor == 0 {
            break
        }
    }
    
    sort.Slice(users, func(i, j int) bool {
        return users[i].Id < users[j].Id
    })
    
    return users
}
