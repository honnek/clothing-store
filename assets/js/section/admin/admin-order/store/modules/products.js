import {concatUrlByParams} from "../../../../../utils/url-generator";
import axios from "axios";
import {StatusCodes} from "http-status-codes";
import {apiContent} from "../../../../../utils/settings";

const state = () => ({
    categories: [],
    staticStore: {
        orderId: window.staticStore.orderId,
        orderProducts: window.staticStore.orderProducts,

        url: {
            viewProduct: window.staticStore.urlViewProduct,
            apiOrderProduct: window.staticStore.urlApiOrderProduct,
            apiCategory: window.staticStore.urlApiCategory,
        }
    },
});

const getters = {};

const actions = {
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
    }
};

export default {
    namespaced: true,
    state,
    getters,
    actions,
    mutations
};