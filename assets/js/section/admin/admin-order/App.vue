<template>
  <div class="table-additional-selection">
    <order-product-add

    />
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
import {mapActions, mapState} from "vuex";
import OrderProductItem from "./components/orderProductItem.vue";
import OrderProductAdd from "./components/orderProductAdd.vue";

export default {
  components: {OrderProductAdd, OrderProductItem},
  created() {
    this.getCategories()
  },
  computed: {
    ...mapState("products", ["staticStore"]),
    productsCount: () => {
      return 123;
    }
  },
  methods: {
    ...mapActions("products", ["getCategories"])
  }
}
</script>