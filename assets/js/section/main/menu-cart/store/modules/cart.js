import axios from "axios";
import {StatusCodes} from "http-status-codes";
import {apiContent, apiConfigPatch} from "../../../../../utils/settings";
import {concatUrlByParams} from "../../../../../utils/url-generator";

const state = () => ({
    cart: {},

    staticStore: {
        url: {
            apiCart: window.staticStore.urlCart,
                viewProduct: window.staticStore.urlViewProduct,
                assetImageProducts: window.staticStore.urlAssetImageProducts,
                apiCartProduct: window.staticStore.urlCartProduct,
                viewCart: window.staticStore.urlViewCart
        }
    }
});


const getters = {
    totalPrice(state) {
        let result = 0

        if (!state.cart.cartProducts) {
            return 0;
        }

        state.cart.cartProducts.forEach(
            cartProduct => {
                result += cartProduct.product.price * cartProduct.quantity
            }
        )

        return result
    }
};


const actions = {
    async getCart({state, commit}) {
        const url = state.staticStore.url.apiCart
        const result = await axios.get(url, apiContent)

        if (result.data && result.data["hydra:member"].length && result.status === StatusCodes.OK) {
            commit("setCart", result.data["hydra:member"][0])
        } else {
            commit("setAlert", {
                type: 'info',
                message: 'Your cart empty!'
            })
        }
    },
    async cleanCart({state, commit}) {
        const url = concatUrlByParams(state.staticStore.url.apiCart, state.cart.id)
        const result = await axios.delete(url, apiContent)

        if (result.status === StatusCodes.NO_CONTENT) {
            commit("setCart", {})
        }
    },
    async removeCartProduct({state, dispatch, commit}, cartProductId) {
        const url = concatUrlByParams(state.staticStore.url.apiCartProduct, cartProductId)
        const result = await axios.delete(url, apiContent)

        if (result.status === StatusCodes.NO_CONTENT) {
            dispatch("getCart")
        }
    }
};

const mutations = {
    setCart(state, cart) {
        state.cart = cart
    }
};

export default {
    namespaced: true, state, getters, actions, mutations
};
