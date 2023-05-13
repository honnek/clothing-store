export function getProductInformativeTitle(product) {
    if (typeof(product.quantity) == "undefined" && typeof(product.quality) != "undefined"){
        product.quantity = product.quality
    }

    return ("#" + product.id + ' / ' + product.title + ' /  $' + product.price + ' / штук: ' + product.quantity)
}
