<template>
  <div class="page-wrap">
    <form-deco />
    <v-snackbar
      :timeout="4000"
      top
      right
      v-model="snackbar"
      class="notification"
      :color="color"
    >
      <div class="action">
        {{ message }}
      </div>
      <v-btn
        text
        icon
        @click="snackbar = false"
      >
        <v-icon>mdi-close</v-icon>
      </v-btn>
    </v-snackbar>
    <v-container class="inner-wrap max-md">
      <div class="full-form-wrap">
        <div class="text-center">
          <h3 class="use-text-title use-text-primary pb-3 text-center">
            {{ $t('common.subscribe_title') }}
          </h3>
        </div>
        <div class="form">
          <v-form
            ref="form"
            v-model="valid"
          >
            <v-row class="spacing6">
              <v-col cols="12" sm="6" class="px-6">
                <v-text-field
                  v-model="name"
                  :rules="nameRules"
                  :label="$t('common.form_name')"
                  color="white"
                  required
                />
              </v-col>
              <v-col cols="12" sm="6" class="px-6">
                <v-text-field
                  v-model="email"
                  :rules="emailRules"
                  :label="$t('common.form_email')"
                  color="white"
                  required
                />
              </v-col>
              
            </v-row>
            <div class="btn-area">
              <v-btn
                :block="isMobile"
                color="primary"
                @click="validate"
                large
              >
                {{ $t('common.subscribe_send') }}
              </v-btn>
            </div>
          </v-form>
        </div>
      </div>
    </v-container>
  </div>
</template>

<style lang="scss" scoped>
@import './form-style.scss';
</style>

<script>
import FormDeco from '../Decoration/FormDeco'
import Hidden from '../Hidden'
import axios from "axios"
export default {
  components: {
    Hidden,
    FormDeco
  },
  data() {
    return {
      valid: true,
      snackbar: false,
      name: '',
      nameRules: [v => !!v || 'Name is required'],
      email: '',
      emailRules: [
        v => !!v || 'E-mail is required',
        v => /.+@.+\..+/.test(v) || 'E-mail must be valid'
      ],
      color: "",
      message: ""
    }
  },
  methods: {
    validate() {
      if (this.$refs.form.validate()) {
        
        axios.post(`/api/newsletter`, { email: this.email, name: this.name })
        .then(res=>{
            if(res.data == "inserted"){
                this.snackbar = true
                this.color = "green"
                this.message = "Welcome aboard!!"
            }else if(res.data == "email adress already exists"){
                this.snackbar = true
                this.color = "orange"
                this.message = "Already subscribed"
            }
        }).catch(err=> {
          this.snackbar = true
                this.color = "red"
                this.message = "Server Error: Please try again later"
        })
        
      }
    }
  },
  computed: {
    isMobile() {
      const xsDown = this.$store.state.breakpoints.xsDown
      return xsDown.indexOf(this.$mq) > -1
    }
  }
}
</script>
