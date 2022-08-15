<template>
<div>
</div>
</template>

<script>
import axios from 'axios'
import https from 'https'
import * as Cookie from 'js-cookie'


 export default({
    created(){
        if(this.$route.query.state == "state"){
            axios.post('/api/rancher',{code: this.$route.query.code}).then(res=>{
            if(res.data.token){
                Cookie.set('rancher_token', res.data.token)
                this.$router.push("/en/dashboard")
            }
            else this.$router.push("/login?err=token_not_found")
        }).then(err=> console.error(err))
        }else{
            window.location = `https://localhost:444/verify-auth?code=${this.$route.query.code}&state=${this.$route.query.state}&session_state=${this.$route.query.session_state}`
        }
    }
   /* async fetch(context){
                        console.log(this.$cookies)
        if(context.query.state == "state"){
            //send request to rancher
            const httpsAgent = new https.Agent({
                rejectUnauthorized: false,
            })
              axios.defaults.httpsAgent = httpsAgent
            axios.post("https://localhost:444/v3-public/keyCloakOIDCProviders/keycloakoidc?action=login", {
                code: context.query.code,
                description: "for website",
                responseType: "token"
            }).then(res=>{
                console.log(res.data)
            })
        }else{
           context.redirect(`https://localhost:444/verify-auth?code=${context.query.code}&state=${context.query.state}&session_state=${context.query.session_state}`)
        }
        
    }*/
 })
</script>