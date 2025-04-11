<?php

namespace App\Controller\Main;

use App\Entity\Category;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class CategoryController extends AbstractController
{
    /**
     * @param Category $category
     *
     * @return Response
     */
    #[Route('/category/{slug}', name: 'main_category_show')]
    public function show(Category $category): Response
    {
        return $this->render('main/category/show.html.twig', [
            'category' => $category,
            'products' => $category->getProducts()->getValues(),
        ]);
    }
}
