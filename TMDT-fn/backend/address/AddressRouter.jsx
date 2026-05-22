const express = require('express');
const router = express.Router();
const axios = require('axios');
const MAP_API = process.env.MAP_API;
let cachedProvinces = null;

router.get('/provinces', async (req, res) => {
    try {
        const response = await axios.get(`${MAP_API}/p/`);
        cachedProvinces = response.data;
        res.json(response.data);
    } catch (error) {
        console.error("province error:", error);
        res.status(500).json({ message: "Lỗi lấy danh sách tỉnh" });
    }
});

router.get('/districts/:provinceCode', async (req, res) => {
    try {
        const { provinceCode } = req.params;
        const response = await axios.get(`${MAP_API}/p/${provinceCode}?depth=2`);
        console.log(`${MAP_API}/p/${provinceCode}?depth=2`)
        res.json(response.data.wards|| []);
    } catch (error) {
        console.error("Distric error:", error);
        res.status(500).json({ message: "Lỗi lấy danh sách huyện" });
    }
});

router.get('/wards/:provinceCode', async (req, res) => {
    try {
        const { provinceCode } = req.params;
        const response = await axios.get(`${MAP_API}/p/${provinceCode}?depth=2`);
        res.json(response.data.wards || []);
    } catch (error) {
        console.error("Ward error:", error);
        res.status(500).json({ message: "Lỗi lấy danh sách xã" });
    }
});

router.get('/reverse-geocode', async (req, res) => {
    try {
        const { lat, lon } = req.query;
        if (!lat || !lon) return res.status(400).json({ message: "Thiếu tọa độ" });

        const osmResponse = await axios.get(process.env.LOCATION_API, {
            params: { format: 'json', lat, lon, 'accept-language': 'vi' },
            headers: { 'User-Agent': 'CellphoneX_Clone_App' }
        });

        const data = osmResponse.data;
        // console.log(osmResponse.data);
        if (!data || !data.display_name) return res.json({ message: "Không tìm thấy địa chỉ" });

        const fullAddress = data.display_name;
        // console.log("cacheProvinces: ",cachedProvinces);
        // const allProvincesRes = await axios.get(`${MAP_API}/p/`);
        const allProvinces = cachedProvinces; //allProvincesRes.data;

        const matchedProvince = allProvinces
            .sort((a, b) => b.name.length - a.name.length)
            .find(p => {
                const provinceName = p.name.toLowerCase();
                const addressLower = fullAddress.toLowerCase();
                const shortName = provinceName
                    .replace(/^(tỉnh|thành phố)\s+/i, "")
                    .trim();
                return addressLower.includes(provinceName) || addressLower.includes(shortName);
            });

        if (!matchedProvince) {
            console.log("Không tìm thấy tỉnh cho địa chỉ:", fullAddress);
        }

        const provinceDetailRes = await axios.get(`${MAP_API}/p/${matchedProvince.code}?depth=2`);
        const provinceData = provinceDetailRes.data;
        //
        const wards = provinceData.wards|| [];
        let detectedWard = "";
        let detailAddress = fullAddress;
        const sortedWards = wards.sort((a, b) => b.name.length - a.name.length);

        for (const ward of sortedWards) {
            // console.log("ward: ",ward);
            if (fullAddress.includes(ward.name)) {
                detectedWard = ward.name;
                const parts = fullAddress.split(ward.name);
                detailAddress = parts[0].trim().replace(/,$/, "").trim();
                break;
            }
        }

        console.log(provinceData.name, detectedWard, detailAddress)

        res.json({
            province: provinceData.name,
            ward: detectedWard || "Không xác định",
            detail: detailAddress || fullAddress,
        });

    } catch (error) {
        console.error("Geocode error:", error);
        res.status(500).json({ message: "Lỗi xử lý dữ liệu địa chỉ" });
    }
});

module.exports = router;