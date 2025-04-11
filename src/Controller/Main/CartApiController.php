<?php

namespace App\Controller\Main;

use App\Entity\Cart;
use App\Entity\CartProduct;
use App\Repository\CartProductRepository;
use App\Repository\CartRepository;
use App\Repository\ProductRepository;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

#[Route('/api', name: 'main_api_')]
class CartApiController extends AbstractController
{
    #[Route('/cart', name: 'cart_save')]
    public function saveCart(
        Request                $request,
        CartProductRepository  $cartProductRepository,
        ProductRepository      $productRepository,
        EntityManagerInterface $entityManager,
        CartRepository         $cartRepository,
    )
    {
        $productId = $request->request->get('productId');
        $sessionId = $request->cookies->get('PHPSESSID');

        $product = $productRepository->findOneBy(['uuid' => $productId]);
        $cart = $cartRepository->findOneBy(['sessionId' => $sessionId]);
        if (!$cart) {
            $cart = new Cart();
            $cart->setSessionId($sessionId);
        }

        $cartProduct = $cartProductRepository->findOneBy(['cart' => $cart, 'product' => $product]);
        if (!$cartProduct) {
            /** @todo вынести в сервис */
            $cartProduct = new CartProduct();
            $cartProduct->setQuantity(1);
            $cartProduct->setCart($cart);
            $cartProduct->setProduct($product);

            $cart->addCartProduct($cartProduct);
        } else {
            $quantity = $cartProduct->getQuantity();
            $cartProduct->setQuantity($quantity + 1);
        }

        $entityManager->persist($cart);
        $entityManager->persist($cartProduct);
        $entityManager->flush();
    }
}
