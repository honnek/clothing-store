<?php

namespace App\Controller\Admin;

use App\Entity\Product;
use App\Form\Admin\EditProductFormType;
use App\Form\DTO\EditProductModel;
use App\Form\Handler\ProductFormHandler;
use App\Repository\ProductRepository;
use App\Utils\Manager\ProductManager;
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
    #[Route('/edit/', name: 'edit_blank')]
    #[Route('/add', name: 'add')]
    public function edit(
        Request            $request,
        ProductFormHandler $productFormHandler,
        Product            $product = null
    ): Response
    {
        $editProductModel = EditProductModel::makeFromProduct($product);
        $form = $this->createForm(EditProductFormType::class, $editProductModel);
        $form->handleRequest($request);

        if ($form->isSubmitted() && $form->isValid()) {

            $product = $productFormHandler->processEditForm($editProductModel, $form);
            $this->addFlash('success', 'The changes have been saved');

            return $this->redirectToRoute('admin_product_edit', ['id' => $product->getId()]);
        }

        if ($form->isSubmitted() && !$form->isValid()) {
            $this->addFlash('warning', 'Something went wrong');
        }

        return $this->render('admin/product/edit.html.twig', [
            'images' => $product?->getProductImages()?->getValues(),
            'product' => $product,
            'form' => $form->createView(),
        ]);
    }

    #[Route('/delete/{id}', name: 'delete')]
    public function delete(Product $product, ProductManager $productManager): Response
    {
        $productManager->remove($product);

        $this->addFlash('warning', 'Product  ' . $product->getTitle() . ' was deleted!');

        return $this->redirectToRoute('admin_product_list');
    }
}
