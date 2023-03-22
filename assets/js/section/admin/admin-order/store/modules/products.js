import {concatUrlByParams, getUrlProductsByCategory} from "../../../../../utils/url-generator";
import axios from "axios";
import {StatusCodes} from "http-status-codes";
import {apiContent} from "../../../../../utils/settings";

const state = () => ({
    categories: [],
    newOrderProduct: {
        categoryId: "",
        productId: "",
        quantity: "",
        pricePerOne: ""
    },
    staticStore: {
        orderId: window.staticStore.orderId,
        orderProducts: window.staticStore.orderProducts,

        url: {
            viewProduct: window.staticStore.urlViewProduct,
            apiOrderProduct: window.staticStore.urlApiOrderProduct,
            apiCategory: window.staticStore.urlApiCategory,
            apiProduct: window.staticStore.urlApiProduct,
        }
    },
    viewProductCountLimit: 25,
});

const getters = {};

const actions = {
    async getProductsByCategory({commit, state}) {
        const url = getUrlProductsByCategory(
            state.staticStore.url.apiProduct,
            state.newOrderProduct.categoryId,
            1,
            state.viewProductCountLimit
        )

        const result = await axios.get(url, apiContent)
        console.log(url, result)
    },
    async getCategories({commit, state}) {
        const url = state.staticStore.url.apiCategory;

        const result = await axios.get(url, apiContent)

        if (result.data && result.status === StatusCodes.OK) {
            commit("setCategories", result.data["hydra:member"])
        }
    },
    async removeOrderProduct({state, dispatch}, orderProductId) {
        const url = concatUrlByParams(state.staticStore.url.apiOrderProduct, orderProductId)

        const result = await axios.delete(url, apiContent)

        if (result.status === StatusCodes.NO_CONTENT) {
            console.log("Deleted!!!")
        }
    }
};

const mutations = {
    setCategories(state, categories) {
        state.categories = categories
    },
    setNewProductInfo(state, formData) {
        state.newOrderProduct.categoryId = formData.category
        state.newOrderProduct.productId = formData.productId
        state.newOrderProduct.quantity = formData.quantity
        state.newOrderProduct.pricePerOne = formData.pricePerOne
    }
};

export default {
    namespaced: true,
    state,
    getters,
    actions,
    mutations
};