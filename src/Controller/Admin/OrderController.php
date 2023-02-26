<?php

namespace App\Controller\Admin;

use App\Entity\Order;
use App\Entity\StaticStorage\OrderStatus;
use App\Repository\OrderRepository;
use App\Utils\Manager\OrderManager;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

#[Route('/admin/order', name: 'admin_order_')]
class OrderController extends AbstractController
{
    #[Route('/list', name: 'list')]
    public function list(OrderRepository $orderRepository): Response
    {
        return $this->render('admin/order/list.html.twig', [
            'orders' => $orderRepository->findBy(['isDeleted' => false], ['id' => 'DESC']),
            'statuses' => OrderStatus::getList(),
        ]);
    }

    #[Route('/edit/{id}', name: 'edit')]
    #[Route('/add', name: 'add')]
    public function edit(
        Request $request,
        Order   $order = null,
    ): Response
    {
        return $this->render('admin/order/edit.html.twig', [
            'order' => $order,
//            'form' => $form->createView(),
        ]);
    }

    #[Route('/delete/{id}', name: 'delete')]
    public function delete(Order $order, OrderManager $orderManager): Response
    {
        $orderManager->remove($order);

        $this->addFlash('warning', 'Заказ  ' . $order->getId() . ' удален!');

        return $this->redirectToRoute('admin_order_list');
    }
}
