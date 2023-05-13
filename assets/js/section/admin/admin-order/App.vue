<template>
  <div class="table-additional-selection">
    <order-product-add

    />
    <hr/>
    <orderProductItem
        v-for="(orderProduct, index) in orderProducts"
        :key="orderProduct.id"
        :order-product="orderProduct"
        :index="index"
    />
    <hr/>

    <total-price-block/>
    <hr/>

  </div>
</template>

<script>
import {mapActions, mapState} from "vuex";
import OrderProductItem from "./components/orderProductItem.vue";
import OrderProductAdd from "./components/orderProductAdd.vue";
import TotalPriceBlock from "./components/totalPriceBlock.vue";

export default {
  components: {TotalPriceBlock, OrderProductAdd, OrderProductItem},
  created() {
    this.getCategories()
    this.getOrderProducts()
  },
  computed: {
    ...mapState("products", ["orderProducts"]),
    productsCount: () => {
      return 123;
    }
  },
  methods: {
    ...mapActions("products", ["getCategories", "getOrderProducts"])
  }
}
</script>