import {concatUrlByParams, getUrlProductsByCategory} from "../../../../../utils/url-generator";
import axios from "axios";
import {StatusCodes} from "http-status-codes";
import {apiContent} from "../../../../../utils/settings";

const state = () => ({
    categories: [],
    categoryProducts: [],
    orderProducts: [],
    busyProductsIds: [],
    newOrderProduct: {
        categoryId: "",
        productId: "",
        quantity: "",
        pricePerOne: ""
    },
    staticStore: {
        orderId: window.staticStore.orderId,

        url: {
            viewProduct: window.staticStore.urlViewProduct,
            apiOrderProduct: window.staticStore.urlApiOrderProduct,
            apiCategory: window.staticStore.urlApiCategory,
            apiProduct: window.staticStore.urlApiProduct,
            apiOrderProductPost: window.staticStore.urlApiOrderProductPost,
            apiOrder: window.staticStore.urlApiOrder,
        }
    },
    viewProductCountLimit: 25,
});

const getters = {
    freeCategoryProducts(state) {
        return state.categoryProducts.filter(
            item => state.busyProductsIds.indexOf(item.id) === -1
        )
    }
};

const actions = {
    async getOrderProducts({commit, state}) {
        const url = concatUrlByParams(state.staticStore.url.apiOrder)

        const result = await axios.get(url, apiContent)
        if (result.data && result.status === StatusCodes.OK) {
            commit("setOrderProducts", result.data.orderProducts)
            commit("setBusyProductsIds")
        }
    },
    async getProductsByCategory({commit, state}) {
        const url = getUrlProductsByCategory(
            state.staticStore.url.apiProduct,
            state.newOrderProduct.categoryId,
            1,
            state.viewProductCountLimit
        )

        const result = await axios.get(url, apiContent)

        if (result.data && result.status === StatusCodes.OK) {
            commit("setCategoryProducts", result.data["hydra:member"])
        }
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
            dispatch('getOrderProducts')
        }
    },
    async addNewOrderProduct({state, dispatch}) {
        const url = state.staticStore.url.apiOrderProductPost
        const data = {
            pricePerOne: state.newOrderProduct.pricePerOne,
            quantity: parseInt(state.newOrderProduct.quantity),
            product: "api/products/" + state.newOrderProduct.productId,
            appOrder: "api/orders/" + state.staticStore.orderId,
        }

        const result = await axios.post(url, data, apiContent)

        if (result.status === StatusCodes.CREATED) {
            dispatch('getOrderProducts')
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
    },
    setCategoryProducts(state, categoryProducts) {
        state.categoryProducts = categoryProducts
    },
    setOrderProducts(state, orderProducts) {
        state.orderProducts = orderProducts
    },
    setBusyProductsIds(state) {
        state.busyProductsIds = state.orderProducts.map(item => item.product.id)
    }
};

export default {
    namespaced: true,
    state,
    getters,
    actions,
    mutations
};