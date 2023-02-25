<?php

namespace App\Controller\Main;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

#[Route('/api', name: 'main_api_')]
class CartApiController extends AbstractController
{
    #[Route('/cart', name: 'cart_save', methods: 'POST')]
    public function saveCart(Request $request): Response
    {
        $productId = $request->request->get('productId');
        $sessionId = $request->cookies->get('PHPSESSID');
        dd($productId, $sessionId);
        return new JsonResponse([
           'success' => false,
           'data' => [
               'test' => 132
           ],
        ]);
    }
}
