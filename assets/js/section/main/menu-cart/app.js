import vue from "vue";
import App from "./App.vue";
import Vue from "vue";
import store from "./store";

const vueMenuCartInstance =  new Vue({
    el: "#appMainMenuCart",
    store,
    render : h => h(App)
})

window.vueMenuCartInstance = {}
window.vueMenuCartInstance.addCartProduct =
    (productData) => vueMenuCartInstance.$store.dispatch('cart/addCartProduct', productData)