<?php

namespace App\Controller\Main;

use App\Repository\CartRepository;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class CartController extends AbstractController
{
    #[Route('/cart', name: 'main_cart_show')]
    public function index(Request $request, CartRepository $cartRepository): Response
    {
        $sessionId = $request->cookies->get('PHPSESSID');
        $cart = $cartRepository->findOneBy(['sessionId' => $sessionId]);

        return $this->render('main/cart/show.html.twig', [
            'cart' => $cart,
        ]);
    }
}
