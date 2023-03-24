<?php

namespace App\Form\Handler;

use App\Entity\Order;
use App\Utils\Manager\OrderManager;

class OrderFormHandler
{

    private OrderManager $orderManager;

    public function __construct(OrderManager $orderManager)
    {
        $this->orderManager = $orderManager;
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
}
