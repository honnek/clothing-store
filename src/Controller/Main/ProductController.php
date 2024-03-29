<?php

namespace App\Controller\Main;

use Nelmio\ApiDocBundle\Annotation\Model;
use Swagger\Annotations as SWG;
use Symfony\Component\Routing\Annotation\Route;
use Sensio\Bundle\FrameworkExtraBundle\Configuration\Security;
use App\Entity\Product;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Exception\NotFoundHttpException;

class ProductController extends AbstractController
{
    #[Route('/product', name: 'main_product_show_blank')]
    #[Route('/product/{uuid}', name: 'main_product_show')]
    public function show(Product $product = null): Response
    {
        if (!$product) {
            throw new NotFoundHttpException('Не найден данный продукт');
        }

        return $this->render('main/product/show.html.twig', [
            'product' => $product,
        ]);
    }
}
