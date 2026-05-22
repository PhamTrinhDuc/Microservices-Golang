const express = require('express');
const router = express.Router();
const axios = require('axios');
const VTP_API = process.env.VTP_API;

async function getToken(){
    const loginData = {
        USERNAME: process.env.USERNAME_VTP,
        PASSWORD: process.env.PASSWORD_VTP
    };
    let res = await axios.post(`${VTP_API}/user/Login`,loginData);
    const softToken = res.data.data.token;
    console.log(softToken);
    const config = {
        headers: {
            'Token': softToken,
            'Content-Type': 'application/json',
        }
    };
    res = await axios.post(`${VTP_API}/user/ownerconnect`,loginData,config);
    console.log(res.data);
    return res.data.data.token;
    //console.log("Token: ",data.token);
}

router.get("/fee", async (req, res) => {

    const product_weight = 1000;
    const sender_address = "Đại Mỗ, Nam Từ liêm, Hà Nội";
    const receiver_address = "Khu phố 14, Hẻm 129/47 Đường Nguyễn Trãi, Phường Chợ Quán, Thành phố Hồ Chí Minh";
    const product_type = "HH";
    const product_price = 59700;
    const money_collection = "0";
    const product_lenght = 0;
    const product_width = 0;
    const product_height = 0;

    const token = getToken();

    const header  = {
        headers: {
            'Token': token,
            'Content-Type': 'application/json',
        }
    };

    const bodyServices = {
        "SENDER_ADDRESS" : sender_address,
        "RECEIVER_ADDRESS" : receiver_address,
        "PRODUCT_TYPE" : product_type,
        "PRODUCT_WEIGHT" : product_weight,
        "PRODUCT_PRICE" : product_price,
        "MONEY_COLLECTION" : money_collection,
        "PRODUCT_LENGTH" : product_lenght,
        "PRODUCT_WIDTH" : product_width,
        "PRODUCT_HEIGHT" : product_height,
        "TYPE" :1
    }

    const resServices = await axios.post(`${VTP_API}/order/getPriceAllNlp`,bodyServices,header);
    //console.log(resServices.data);
    res.json(resServices.data);
});

module.exports = router;