<?php

namespace App\Controller\Admin;

use App\Entity\Order;
use App\Entity\StaticStorage\OrderStatus;
use App\Form\Admin\EditOrderFormType;
use App\Form\Admin\FilterType\OrderFilterForm;
use App\Form\DTO\EditOrderModel;
use App\Form\Handler\OrderFormHandler;
use App\Utils\Manager\OrderManager;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

#[Route('/admin/order', name: 'admin_order_')]
class OrderController extends AbstractController
{
    #[Route('/list', name: 'list')]
    public function list(
        Request          $request,
        OrderFormHandler $orderFormHandler
    ): Response
    {
        $filterForm = $this->createForm(OrderFilterForm::class, new EditOrderModel())->handleRequest($request);
        $pagination = $orderFormHandler->processOrderFiltersForm($request, $filterForm);

        return $this->render('admin/order/list.html.twig', [
            'pagination' => $pagination,
            'statuses' => OrderStatus::getList(),
            'form' => $filterForm->createView()
        ]);
    }

    #[Route('/edit/{id}', name: 'edit')]
    #[Route('/add', name: 'add')]
    public function edit(
        Request          $request,
        OrderFormHandler $orderFormHandler,
        Order            $order = null,
    ): Response
    {
        if (!$order) {
            $order = new Order();
        }

        $form = $this->createForm(EditOrderFormType::class, $order);
        $form->handleRequest($request);

        if ($form->isSubmitted() && $form->isValid()) {

            $order = $orderFormHandler->processEditForm($order);

            $this->addFlash('success', 'The changes have been saved');

            return $this->redirectToRoute('admin_order_edit', ['id' => $order->getId()]);
        }

        if ($form->isSubmitted() && !$form->isValid()) {
            $this->addFlash('warning', 'Something went wrong');
        }

        return $this->render('admin/order/edit.html.twig', [
            'order' => $order,
            'orderProducts' => [],
            'form' => $form->createView(),
        ]);
    }

    #[Route('/delete/{id}', name: 'delete')]
    public function delete(Order $order, OrderManager $orderManager): Response
    {
        $orderManager->remove($order);

        $this->addFlash('warning', 'Order ' . $order->getId() . ' was deleted!');

        return $this->redirectToRoute('admin_order_list');
    }
}
