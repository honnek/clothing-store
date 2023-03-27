<template>
  <tr>
    <td class="product-col">
      <div class="text-center">
        <figure>
          <a
              :href="urlShowProduct"
              target="_blank"
          >
            <img
                :src="getUrlProductImage(productImage)"
                :alt="cartProduct.product.title"
            >
          </a>
        </figure>
        <div class="product-title">
          <a
              href="urlShowProduct"
              target="_blank"
          >
            {{ cartProduct.product.tittle }}
          </a>
        </div>
      </div>
    </td>
    <td class="price-col">
      ${{ cartProduct.product.price }}
    </td>
    <td class="quantity-col">
      <input
          v-model="quantity"
          type="number"
          class="form-control"
          min="1"
          step="1"
      >
    </td>
    <td class="total-col">
      ${{ productPrice }}
    </td>

    <td class="remove-col">
      <a
          href="#"
          class="btn-remove"
          title="Remove product"
          @click="removeCartProduct(cartProduct.id)"
      >
        X
      </a>
    </td>
  </tr>
</template>

<script>
import {mapActions, mapState} from "vuex";

export default {
  name: "CartProductItem",
  data() {
    return {
      quantity: 1,
    }
  },
  props: {
    cartProduct: {
      type: Object,
      default: () => {
      }
    },
  },
  computed: {
    ...mapState("cart", ["staticStore"]),
    productImage() {
      const productImages = this.cartProduct.product.productImages

      return productImages.length ? productImages[0] : null;
    },
    productPrice() {
      return this.quantity * this.cartProduct.product.price
    },
    urlShowProduct() {
      return this.staticStore.url.viewProduct + "/" + this.cartProduct.product.uuid
    }
  },
  methods: {
    ...mapActions("cart", ["removeCartProduct"]),
    getUrlProductImage(productImage) {
      return (
          this.staticStore.url.assetImageProducts + "/" +
          this.cartProduct.product.id + "/" +
          productImage.filenameSmall
      )
    },
    removeCartProduct() {
      return true
    }
  }
}
</script>

<style scoped>

</style>