<?php

namespace App\Controller\Main;

use App\Entity\Product;
use App\Form\Admin\EditProductFormType;
use Doctrine\Persistence\ManagerRegistry;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class DefaultController extends AbstractController
{
    /**
     * @return Response
     */
    #[Route('/', name: 'main_homepage')]
    public function index(ManagerRegistry $doctrine): Response
    {
        return $this->render('main/default/index.html.twig');
    }

    /**
     * @param Request $request
     * @param ManagerRegistry $doctrine
     * @param int|null $id
     * @return Response
     */
    #[Route('/edit-product/{id}', name: 'product_edit', requirements: ["id" => "\d+"])]
    #[Route('/add-product', name: 'product_add')]
    public function editProduct(Request $request, ManagerRegistry $doctrine, int $id = null): Response
    {
        $entityManager = $doctrine->getManager();

        if ($id) {
            $product = $entityManager->getRepository(Product::class)->find($id);
        } else {
            $product = new Product();
        }

        $form = $this->createForm(EditProductFormType::class, $product);
        /** Присваиваем поля из Request в форму */
        $form->handleRequest($request);

        if ($form->isSubmitted() && $form->isValid()) {

            /** Сохраним модель в бд */
            $entityManager->persist($product);
            /** Обновим бд */
            $entityManager->flush();

            /** После сохранения продукта редирект на текущий продукт */
            return $this->redirectToRoute('product_edit', ['id' => $product->getId()]);
        }

        return $this->render('main/default/edit_product.html.twig', [
            'form' => $form->createView(),
        ]);
    }
}
