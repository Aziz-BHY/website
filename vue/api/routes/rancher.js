const { Router } = require('express')
const router = Router()
const axios = require('axios')
const https = require('https')
router.post('/rancher', function(req, res) {
    const httpsAgent = new https.Agent({
        rejectUnauthorized: false,
    })
    axios.defaults.httpsAgent = httpsAgent
    axios
    .post("https://localhost:444/v3-public/keyCloakOIDCProviders/keycloakoidc?action=login", {
        code: req.body.code,
        description: "for website",
        responseType: "token"
    })
    .then(response => {
      res.json(response.data)
    })
    .catch(error => {
        console.log(error)
      res.status(500).json(error.data)
    })
})

module.exports = router
