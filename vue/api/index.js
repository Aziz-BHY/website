const express = require('express')
const app = express()
const bodyParser = require('body-parser')

app.use(bodyParser.json())
app.use(bodyParser.urlencoded({ extended: true }))

const eventRouter = require('./routes/events.js')
const paymeeRouter = require('./routes/paymee.js')
const newsletterRouter = require('./routes/newsletter.js')
const mailRouter = require('./routes/mail.js')
const loginRouter = require('./routes/login.js')
const registerRouter = require('./routes/register.js')
const projectRouter = require('./routes/project.js')
const rancherRouter = require('./routes/rancher.js')

app.use(eventRouter)
app.use(paymeeRouter)
app.use(newsletterRouter)
app.use(mailRouter)
app.use(loginRouter)
app.use(registerRouter)
app.use(projectRouter)
app.use(rancherRouter)

if (require.main === module) {
  const port = 3001
  app.listen(port, () => {
    console.log(`API server listening on port ${port}`)
  })
}

module.exports = { path: '/api', handler: app }
