<template>
  <div class="table-additional-selection">
    <hr/>
      <orderProductItem
        v-for="(orderProduct, index) in staticStore.orderProducts"
        :key="orderProduct.id"
        :order-product="orderProduct"
        :index="index"
        />
    <hr/>

  </div>
</template>

<script>
import {mapState} from "vuex";
import OrderProductItem from "./components/orderProductItem.vue";

export default {
  components: {OrderProductItem},
  created() {
    console.log(window.staticStore)
  },
  computed: {
    ...mapState("products", ["staticStore"]),
    productsCount: () => {
      return 123;
    }
  }
}
</script>