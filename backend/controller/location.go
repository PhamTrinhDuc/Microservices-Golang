package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type LocationController struct {
	client *http.Client

	// Cache in-memory
	cacheMu      sync.RWMutex
	provinces    []Province
	provinceTime time.Time

	districts    map[int][]District
	districtTime map[int]time.Time

	wards    map[int][]Ward
	wardTime map[int]time.Time
}

type Province struct {
	Name     string `json:"name"`
	Code     int    `json:"code"`
	Codename string `json:"codename"`
}

type District struct {
	Name     string `json:"name"`
	Code     int    `json:"code"`
	Codename string `json:"codename"`
}

type ProvinceDetail struct {
	Province
	Districts []District `json:"districts"`
}

type Ward struct {
	Name     string `json:"name"`
	Code     int    `json:"code"`
	Codename string `json:"codename"`
}

type DistrictDetail struct {
	District
	Wards []Ward `json:"wards"`
}

func NewLocationController() *LocationController {
	return &LocationController{
		client:       &http.Client{Timeout: 10 * time.Second},
		districts:    make(map[int][]District),
		districtTime: make(map[int]time.Time),
		wards:        make(map[int][]Ward),
		wardTime:     make(map[int]time.Time),
	}
}

// GetProvinces handles GET /location/provinces.
func (lc *LocationController) GetProvinces(ctx *gin.Context) {
	lc.cacheMu.RLock()
	if len(lc.provinces) > 0 && time.Since(lc.provinceTime) < 24*time.Hour {
		provinces := lc.provinces
		lc.cacheMu.RUnlock()
		utils.RespondOK(ctx, provinces)
		return
	}
	lc.cacheMu.RUnlock()

	lc.cacheMu.Lock()
	defer lc.cacheMu.Unlock()

	// Double check
	if len(lc.provinces) > 0 && time.Since(lc.provinceTime) < 24*time.Hour {
		utils.RespondOK(ctx, lc.provinces)
		return
	}

	resp, err := lc.client.Get(utils.GetEnvString("ADDRESS_API", "https://provinces.open-api.vn/api/v2") + "/p/")
	if err != nil {
		utils.RespondInternalError(ctx, fmt.Sprintf("Failed to fetch provinces: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.RespondInternalError(ctx, fmt.Sprintf("API returned status: %d", resp.StatusCode))
		return
	}

	var provinces []Province
	if err := json.NewDecoder(resp.Body).Decode(&provinces); err != nil {
		utils.RespondInternalError(ctx, fmt.Sprintf("Failed to decode provinces: %v", err))
		return
	}

	lc.provinces = provinces
	lc.provinceTime = time.Now()

	utils.RespondOK(ctx, provinces)
}

// GetDistricts handles GET /location/provinces/:code/districts.
func (lc *LocationController) GetDistricts(ctx *gin.Context) {
	codeStr := ctx.Param("code")
	provinceCode, err := strconv.Atoi(codeStr)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid province code")
		return
	}

	lc.cacheMu.RLock()
	if val, ok := lc.districts[provinceCode]; ok && time.Since(lc.districtTime[provinceCode]) < 24*time.Hour {
		lc.cacheMu.RUnlock()
		utils.RespondOK(ctx, val)
		return
	}
	lc.cacheMu.RUnlock()

	lc.cacheMu.Lock()
	defer lc.cacheMu.Unlock()

	// Double check
	if val, ok := lc.districts[provinceCode]; ok && time.Since(lc.districtTime[provinceCode]) < 24*time.Hour {
		utils.RespondOK(ctx, val)
		return
	}

	url := fmt.Sprintf("https://provinces.open-api.vn/api/v2/p/%d?depth=2", provinceCode)
	resp, err := lc.client.Get(url)
	if err != nil {
		utils.RespondInternalError(ctx, fmt.Sprintf("Failed to fetch districts: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.RespondInternalError(ctx, fmt.Sprintf("API returned status: %d", resp.StatusCode))
		return
	}

	var detail ProvinceDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		utils.RespondInternalError(ctx, fmt.Sprintf("Failed to decode districts: %v", err))
		return
	}

	lc.districts[provinceCode] = detail.Districts
	lc.districtTime[provinceCode] = time.Now()

	utils.RespondOK(ctx, detail.Districts)
}

// GetWards handles GET /location/districts/:code/wards.
func (lc *LocationController) GetWards(ctx *gin.Context) {
	codeStr := ctx.Param("code")
	districtCode, err := strconv.Atoi(codeStr)
	if err != nil {
		utils.RespondBadRequest(ctx, "invalid district code")
		return
	}

	lc.cacheMu.RLock()
	if val, ok := lc.wards[districtCode]; ok && time.Since(lc.wardTime[districtCode]) < 24*time.Hour {
		lc.cacheMu.RUnlock()
		utils.RespondOK(ctx, val)
		return
	}
	lc.cacheMu.RUnlock()

	lc.cacheMu.Lock()
	defer lc.cacheMu.Unlock()

	// Double check
	if val, ok := lc.wards[districtCode]; ok && time.Since(lc.wardTime[districtCode]) < 24*time.Hour {
		utils.RespondOK(ctx, val)
		return
	}

	url := fmt.Sprintf("%s/d/%d?depth=2", utils.GetEnvString("ADDRESS_API", "https://provinces.open-api.vn/api/v2"), districtCode)
	resp, err := lc.client.Get(url)
	if err != nil {
		utils.RespondInternalError(ctx, fmt.Sprintf("Failed to fetch wards: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.RespondInternalError(ctx, fmt.Sprintf("API returned status: %d", resp.StatusCode))
		return
	}

	var detail DistrictDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		utils.RespondInternalError(ctx, fmt.Sprintf("Failed to decode wards: %v", err))
		return
	}

	lc.wards[districtCode] = detail.Wards
	lc.wardTime[districtCode] = time.Now()

	utils.RespondOK(ctx, detail.Wards)
}
