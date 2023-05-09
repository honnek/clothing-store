<?php

namespace App\Form\Handler;

use App\Entity\Order;
use App\Utils\Manager\OrderManager;
use Knp\Component\Pager\PaginatorInterface;
use Symfony\Component\HttpFoundation\Request;

class OrderFormHandler
{

    private PaginatorInterface $paginator;
    private OrderManager $orderManager;

    public function __construct(OrderManager $orderManager, PaginatorInterface $paginator)
    {
        $this->orderManager = $orderManager;
        $this->paginator = $paginator;
    }

    /**
     * @param Order $order
     * @return Order
     */
    public function processEditForm(Order $order): Order
    {
        $this->orderManager->recalculateOrderTotalPrice($order);
        $this->orderManager->save($order);

        return $order;
    }

    public function processOrderFiltersForm(Request $request, string $filterForm)
    {
        $orders = $this->orderManager->getRepository()
            ->createQueryBuilder('o')
            ->leftJoin('o.owner', 'u')
            ->where('o.isDeleted = :isDeleted')
            ->setParameter('isDeleted', false)
            ->getQuery();

        return $this->paginator->paginate($orders, $request->get('page', 1));
    }
}
