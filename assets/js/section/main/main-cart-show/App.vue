<template>
  <div class="row">
    <div class="col-lg-12 order-block">
      <div class="order-content">
        <Alert/>
        <div v-if="showCartContent">
          <cart-product-list/>
          <cart-total-price/>
          <a
            class="btn btn-success mb-3 text-white"
            @click="makeOrder"
            >
            Make Order
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import CartProductList from "./components/CartProductList.vue";
import CartTotalPrice from "./components/CartTotalPrice.vue";
import {mapActions} from "vuex";
import Alert from "./components/Alert.vue";

export default {
  name: "App",
  components: {Alert, CartTotalPrice, CartProductList},
  created() {
      return this.getCart()
  },
  computed: {
    showCartContent() {
      return true
    }
  },
  methods: {
    ...mapActions("cart", ["getCart"]),
    makeOrder() {
      return true
    }
  }
}
</script>

<style scoped>

</style>