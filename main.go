package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	baseURL       = "https://api.v2.rainyun.com/user/reward/items"
	xAPIKeyHeader = "x-api-key"
)

type Item struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Points         int    `json:"points"`
	AvailableStock int    `json:"available_stock"`
}

type Response struct {
	Code int    `json:"code"`
	Data []Item `json:"data"`
}

type PurchaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func getAPIKey() string {
	var apiKey string
	fmt.Print("请输入API秘钥: ")
	fmt.Scan(&apiKey)
	return apiKey
}

func getProductList(apiKey string) ([]Item, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(xAPIKeyHeader, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("failed to fetch products, code: %d", response.Code)
	}

	return response.Data, nil
}

func purchaseItem(apiKey string, itemID int) (PurchaseResponse, error) {
	client := &http.Client{}
	postBody, _ := json.Marshal(map[string]int{"item_id": itemID})
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(postBody))
	if err != nil {
		return PurchaseResponse{}, err
	}
	req.Header.Add(xAPIKeyHeader, apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return PurchaseResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PurchaseResponse{}, err
	}

	var response PurchaseResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return PurchaseResponse{}, err
	}

	return response, nil
}

func main() {
	apiKey := getAPIKey()
	for {
		items, err := getProductList(apiKey)
		if err != nil {
			fmt.Println("获取产品列表失败:", err)
			return
		}

		fmt.Println("可兑换产品列表:")
		for _, item := range items {
			fmt.Printf("ID: %d, Name: %s\n", item.ID, item.Name)
		}
		fmt.Print("抢购时输入0再按回车可以终止操作\n")
		fmt.Print("请输入要抢购的产品ID (输入0退出): ")

		var itemID int
		fmt.Scan(&itemID)
		if itemID == 0 {
			break
		}

		stop := make(chan bool)
		go func() {
			for {
				var input int
				fmt.Scan(&input)
				if input == 0 {
					stop <- true
					break
				}
			}
		}()

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				response, err := purchaseItem(apiKey, itemID)
				if err != nil {
					fmt.Println("兑换失败:", err)
					return
				}
				if response.Code == 200 {
					fmt.Println("兑换成功!")
					stop <- true
				} else {
					fmt.Println("兑换尝试:", response.Message)
				}
			case <-stop:
				fmt.Println("操作已终止")
				goto end
			}
		}
	end:
		continue
	}
}