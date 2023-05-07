<template>
    <div class="dropdown cart-dropdown">
        <!-- @TODO en здесь костыль, вместо en должна быть переменная locale -->
        <a href="/en/cart" class="cart-dropdown-btn-toggle">
            <i class="fas fa-shopping-cart"></i>
            <span class="count">{{ countCartProducts }}</span>
        </a>

        <div class="dropdown-menu cart-dropdown-window">
            <CartProductList

            />

            <CartTotalPrice

            />

            <div v-if="countCartProducts">
                <CartActions/>
            </div>
            <div v-else>
                Ваша корзина пуста ...
            </div>

        </div>
    </div>
</template>

<script>
import CartTotalPrice from "./components/CartTotalPrice.vue";
import CartActions from "./components/CartActions.vue";
import CartProductList from "./components/CartProductList.vue";
import {mapActions, mapState} from "vuex";

export default {
    name: "App",
    components: {CartProductList, CartActions, CartTotalPrice},
    created() {
        return this.getCart()
    },
    computed: {
        ...mapState("cart", ["cart"]),
        countCartProducts() {
            return this.cart.cartProducts ? this.cart.cartProducts.length : 0
        }
    },
    methods: {
        ...mapActions("cart", ["getCart"]),
    }
}
</script>

<style scoped>

</style>
