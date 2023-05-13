<template>
  <table class="table table-cart table-mobile main-cart-show">
    <thead>
    <tr>
      <td>Product</td>
      <td>Price</td>
      <td>Quantity</td>
      <td>Total</td>
      <td></td>
    </tr>
    </thead>
    <tbody>
    <cart-product-item
      v-for="cartProduct in cart.cartProducts"
      :key="cartProduct.id"
      :cart-product="cartProduct"
      />
    </tbody>
  </table>
</template>

<script>
import CartProductItem from "./CartProductItem.vue";
import {mapState} from "vuex";

export default {
  name: "CartProductList",
  components: {CartProductItem},
  computed: {
    ...mapState("cart", ["cart"]),
  },
}
</script>

<style scoped>

</style>