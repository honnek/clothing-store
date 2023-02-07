<?php

namespace App\Controller\Admin;

use App\Entity\Product;
use App\Form\EditProductFormType;
use App\Form\Handler\ProductFormHandler;
use App\Repository\ProductRepository;
use Doctrine\Persistence\ManagerRegistry;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

#[Route('/admin/product', name: 'admin_product_')]
class ProductController extends AbstractController
{
    /**
     * @TODO Категория, обложка, пагинатор
     * @param ProductRepository $productRepository
     * @return Response
     */
    #[Route('/list', name: 'list')]
    public function list(ProductRepository $productRepository): Response
    {
        return $this->render('admin/product/list.html.twig', [
            'products' => $productRepository->findBy([
                'isDeleted' => false
            ]),
        ]);
    }

    #[Route('/edit/{id}', name: 'edit')]
    #[Route('/add', name: 'add')]
    public function edit(
        Request            $request,
        ManagerRegistry    $managerRegistry,
        ProductFormHandler $productFormHandler,
        Product            $product = null
    ): Response
    {
        $form = $this->createForm(EditProductFormType::class, $product);

        $form->handleRequest($request);

        if ($form->isSubmitted() && $form->isValid()) {
            $product = $productFormHandler->processEditForm($product, $form);

            return $this->redirectToRoute('admin_product_edit', ['id' => $product->getId()]);
        }

        return $this->render('admin/product/edit.html.twig', [
            'product' => $product,
            'form' => $form->createView(),
        ]);
    }

    #[Route('/delete/{id}', name: 'delete')]
    public function delete(): Response
    {
//        return $this->render('admin/product/index.html.twig', [
//            'controller_name' => 'ProductController',
//        ]);
    }
}
