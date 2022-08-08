const { Router } = require('express')
const router = Router()
const axios = require('axios')

router.put('/events', function(req, res){
  axios.put(`${process.env.SERVER_URL}/events/attendee`, req.body).then(response => {
      res.json(response.data)
  }).catch(err=> res.status(500).json(err))
})

module.exports = router