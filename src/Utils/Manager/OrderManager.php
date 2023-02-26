<?php

namespace App\Utils\Manager;


use App\Entity\Cart;
use App\Entity\Order;
use App\Entity\OrderProduct;
use App\Entity\StaticStorage\OrderStatus;
use App\Entity\User;
use Doctrine\ORM\EntityManagerInterface;
use Doctrine\Persistence\ObjectRepository;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\Security\Core\User\UserInterface;

class OrderManager extends AbstractBaseManager
{
    private CartManager $cartManager;

    public function __construct(EntityManagerInterface $entityManager, CartManager $cartManager)
    {
        parent::__construct($entityManager);
        $this->cartManager = $cartManager;
    }

    /**
     * @return ObjectRepository
     */
    public function getRepository(): ObjectRepository
    {
        return $this->entityManager->getRepository(Order::class);
    }

    /**
     * @param string $sessionId
     * @param User $user
     * @return void
     */
    public function createOrderFromCartBySessionId(string $sessionId, User $user)
    {
        $cart = $this->cartManager->getRepository()->findOneBy(['sessionId' => $sessionId]);

        if ($cart) {
            $this->createOrderFromCart($cart, $user);
        }


    }

    /**
     * @param Cart $cart
     * @param User $user
     */
    public function createOrderFromCart(Cart $cart, User $user)
    {
        $totalPrice = 50;
        $order = new Order();
        $order->setStatus(OrderStatus::CREATED);
        $order->setOwner($user);

        foreach ($cart->getCartProducts() as $cartProduct) {
            $orderProduct = new OrderProduct();
            $orderProduct->setAppOrder($order);
            $orderProduct->setProduct($cartProduct->getProduct());
            $orderProduct->setQuantity($cartProduct->getQuantity());
            $orderProduct->setPricePerOne($cartProduct->getProduct()->getPrice());

            $order->addOrderProduct($orderProduct);

            $totalPrice += $orderProduct->getQuantity() * $orderProduct->getPricePerOne();

            $this->entityManager->persist($orderProduct);
        }

        $order->setTotalPrice($totalPrice);

        $this->entityManager->persist($order);
        $this->entityManager->flush();

        $this->cartManager->remove($cart);
        dd($order);
    }

    /**
     * @param object $order
     * @return void
     */
    public function remove(object $order): void
    {
        /** @var Order $order */
        $order->setIsDeleted(true);

        $this->save($order);
        $this->entityManager->flush();
    }
}
