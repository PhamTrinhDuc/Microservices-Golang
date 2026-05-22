require('dotenv').config();

const generatePaymentCode = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let result = '';
    for (let i = 0; i < 6; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return `CPX${result}`;
};

const generateVietQR = (amount, paymentCode) => {
    const BANK_ID = "TPB";
    const ACCOUNT_NO = "00001221323";
    const TEMPLATE = "compact";

    return `${process.env.QR}/${BANK_ID}-${ACCOUNT_NO}-compact.png?amount=${amount}&addInfo=${encodeURIComponent(paymentCode)}`;
};

module.exports = { generatePaymentCode, generateVietQR };