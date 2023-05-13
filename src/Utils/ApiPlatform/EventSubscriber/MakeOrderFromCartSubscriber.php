<?php

namespace App\Utils\ApiPlatform\EventSubscriber;

use ApiPlatform\Symfony\EventListener\EventPriorities;
use App\Entity\Order;
use App\Entity\StaticStorage\OrderStatus;
use App\Entity\User;
use App\Utils\Manager\OrderManager;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Event\ViewEvent;
use Symfony\Component\HttpKernel\KernelEvents;
use Symfony\Component\Security\Core\Security;

/**
 * Прослойка для отлова post запросов на создание заказа
 *
 *
 * @TODO отрефакторить!
 */
class MakeOrderFromCartSubscriber implements EventSubscriberInterface
{
    public Security $security;
    public OrderManager $orderManager;

    public function __construct(Security $security, OrderManager $orderManager)
    {
        $this->security = $security;
        $this->orderManager = $orderManager;
    }

    public static function getSubscribedEvents()
    {
        return [
            KernelEvents::VIEW => [
                [
                    'makeOrder', EventPriorities::PRE_WRITE
                ]
            ],
        ];
    }

    public function makeOrder(ViewEvent $event)
    {
        $order = $event->getControllerResult();
        $method = $event->getRequest()->getMethod();

        if (!$order instanceof Order || Request::METHOD_POST !== $method) {
            return;
        }

        /** @var User $user */
        $user = $this->security->getUser();

        if (!$user) {
            return;
        }

        $order->setOwner($user);

        $contentJson = $event->getRequest()->getContent();
        if (!$contentJson) {
            return;
        }

        $content = json_decode($contentJson, true);
        if (!array_key_exists('cartId', $content)) {
            return;
        }

        $cartId = (int) $content['cartId'];

        $this->orderManager->addOrderProductsFromCarts($order, $cartId);
        $this->orderManager->recalculateOrderTotalPrice($order);

        $order->setStatus(OrderStatus::CREATED);
    }
}
