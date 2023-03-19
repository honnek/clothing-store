<template>
  <div class="row mb-1">

    <div class="col-mb-8">
      {{ rowNumber }}
    </div>

    <div class="col-mb-1">
      {{ productTitle }}
    </div>

    <div class="col-mb-3">
      {{ categoryTitle }}
    </div>

    <div class="col-mb-2">
      {{ orderProduct.quality }}
    </div>

    <div class="col-mb-2">
      ${{ orderProduct.pricePerOne }}
    </div>

    <div class="col-mb-3">
      <button class="btn btn-outline-info" @click="viewDetails">
        Details
      </button>
      <button class="btn btn-outline-success" @click="remove">
        Remove
      </button>
    </div>

  </div>
</template>

<script>
import products from "../store/modules/products";
import {mapState} from "vuex";
import {getUrlViewProduct} from "../../../../utils/url-generator";

export default {
  name: "orderProductItem",
  props: {
    orderProduct: {
      type: Object,
      default: () => {
      }
    },
    index: {
      type: Number,
      default: 0
    }
  },
  computed: {
    ...mapState("products", ["staticStore"]),
    rowNumber() {
      return this.index + 1;
    },
    productTitle() {
      return this.orderProduct.product.title;
    },
    categoryTitle() {
      return this.orderProduct.product.category.title;
    },
  },
  methods: {
    viewDetails(event) {
      event.preventDefault();

      const url = getUrlViewProduct(this.staticStore.url.viewProduct, this.orderProduct.product.id);

      window.open(url, "_blank").focus()
    }
  }
}
</script>