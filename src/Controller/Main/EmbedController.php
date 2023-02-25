<?php

namespace App\Controller\Main;

use App\Repository\ProductRepository;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;

class EmbedController extends AbstractController
{
    public function showSimilarProducts(ProductRepository $productRepository, int $limit = 6, ?int $categoryId = null): Response
    {
        $params = [];
        if ($categoryId) {
            $params['category'] = $categoryId;
        }

        return $this->render('/main/_embed/similar_products.html.twig', [
            'products' => $productRepository->findBy($params, ['id' => 'DESC'], $limit),
        ]);
    }
}
