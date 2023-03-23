<template>
  <div class="row">

    <div class="col">
      {{ rowNumber }}
    </div>

    <div class="col">
      {{ productTitle }}
    </div>

    <div class="col">
      {{ categoryTitle }}
    </div>

    <div class="col">
      {{ orderProduct.quantity }}
    </div>

    <div class="col">
      ${{ orderProduct.pricePerOne }}
    </div>

    <div class="col">
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
import {mapActions, mapState} from "vuex";
import {getUrlViewProduct} from "../../../../utils/url-generator";
import {getProductInformativeTitle} from "../../../../utils/title-formatter";

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
      return getProductInformativeTitle(this.orderProduct.product);
    },
    categoryTitle() {
      return this.orderProduct.product.category.title;
    },
  },
  methods: {
    ...mapActions("products", ["removeOrderProduct"]),
    viewDetails(event) {
      event.preventDefault();

      const url = getUrlViewProduct(this.staticStore.url.viewProduct, this.orderProduct.product.id);

      window.open(url, "_blank").focus()
    },
    remove(event) {
      event.preventDefault();

      this.removeOrderProduct(this.orderProduct.id);
    }
  }
}
</script>