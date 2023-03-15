<template>
  <div>
    Vue App {{ testProperty }}
  </div>
</template>

<script>
import {mapState} from "vuex";

export default {
  created() {
    console.log(window.staticStore)
  },
  computed: {
    ...mapState("products", ["testProperty"]),
    productsCount: () => {
      return 123;
    }
  }
}
</script>