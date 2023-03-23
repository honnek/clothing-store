<template>
  <div class="row mb-2">

    <div class="col mb-2">
      <select
          v-model="form.category"
          name="add_product_category_select"
          class="form-control"
          @change="getProducts()"
      >

        <option value="disabled"> - choose option -</option>
        <option
            v-for="category in categories"
            :key="category.id"
            :value="category.id"
        >
          {{ category.title }}
        </option>
      </select>
    </div>

    <div v-if="form.category" class="col-mb-3">
      <select
          v-model="form.productId"
          name="add_product_product_select"
          class="form-control"
      >

        <option value="disabled"> - choose option -</option>
        <option
            v-for="categoryProduct in freeCategoryProducts"
            :key="categoryProduct.id"
            :value="categoryProduct.uuid"
        >
          {{ specialProductTitle(categoryProduct) }}
        </option>
      </select>
    </div>

    <div v-if="form.productId" class="col-mb-2">
      <input
          v-model="form.quantity"
          type="number"
          class="form-control"
          placeholder="quantity"
          min="1"
          :max="productQuantityMax"
      />
    </div>

    <div v-if="form.productId" class="col-mb-2">
      <input
          v-model="form.pricePerOne"
          type="number"
          class="form-control"
          placeholder="price per one"
          step="10"
          min="1"
          :max="productPriceMax"
      />
    </div>

    <div v-if="form.productId" class="col-mb-3">
      <button
          class="btn btn-outline-info"
          @click="viewDetails"
      >
        Details
      </button>
      <button
          class="btn btn-outline-success"
          @click="submit"
      >
        Add
      </button>
    </div>
  </div>
</template>

<script>
import {mapActions, mapGetters, mapMutations, mapState} from "vuex";
import {getProductInformativeTitle} from "../../../../utils/title-formatter";
import {getUrlViewProduct} from "../../../../utils/url-generator";

export default {
  name: "orderProductAdd",
  data() {
    return {
      form: {
        categoryId: "",
        productId: "",
        quantity: "",
        pricePerOne: ""
      }
    };
  },
  computed: {
    ...mapState("products", ["categories", "staticStore", "categoryProducts"]),
    ...mapGetters("products", ["freeCategoryProducts"]),
    productQuantityMax() {
      const productData = this.freeCategoryProducts.find(
          product => product.uuid === this.form.productId
      )

      return parseInt(productData.quantity)
    },
    productPriceMax() {
      const productData = this.freeCategoryProducts.find(
          product => product.uuid === this.form.productId
      )

      return parseFloat(productData.price)
    }
  },
  methods: {
    ...mapMutations("products", ["setNewProductInfo"]),
    ...mapActions("products", ["addNewOrderProduct", "getProductsByCategory"]),
    getProducts() {
      this.setNewProductInfo(this.form)
      this.getProductsByCategory()
    },
    specialProductTitle(product) {
      return getProductInformativeTitle(product)
    },
    viewDetails(event) {
      event.preventDefault();

      const url = getUrlViewProduct(this.staticStore.url.viewProduct, this.form.productId);

      window.open(url, "_blank").focus()
    },
    submit(event) {
      event.preventDefault();
      this.setNewProductInfo(this.form)
      this.addNewOrderProduct()
      this.resetFormData()
    },
    resetFormData() {
      Object.assign(this.$data, this.$options.data.apply(this))
    }
  }
}
</script>

<style scoped>

</style>