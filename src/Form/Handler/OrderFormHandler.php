<?php

namespace App\Form\Handler;

use App\Entity\Order;
use App\Utils\Manager\OrderManager;
use Knp\Component\Pager\PaginatorInterface;
use Lexik\Bundle\FormFilterBundle\Filter\FilterBuilderUpdater;
use Symfony\Component\Form\FormInterface;
use Symfony\Component\HttpFoundation\Request;

class OrderFormHandler
{

    private PaginatorInterface $paginator;
    private OrderManager $orderManager;

    private FilterBuilderUpdater $filterBuilderUpdater;

    public function __construct(
        OrderManager $orderManager,
        PaginatorInterface $paginator,
        FilterBuilderUpdater $filterBuilderUpdater,
    )
    {
        $this->orderManager = $orderManager;
        $this->paginator = $paginator;
        $this->filterBuilderUpdater = $filterBuilderUpdater;
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

    public function processOrderFiltersForm(Request $request, FormInterface $filterForm)
    {
        $query = $this->orderManager->getRepository()
            ->createQueryBuilder('o')
            ->leftJoin('o.owner', 'u')
            ->where('o.isDeleted = :isDeleted')
            ->setParameter('isDeleted', false);

        if ($filterForm->isSubmitted()) {
            $this->filterBuilderUpdater->addFilterConditions($filterForm, $query);
        }


        return $this->paginator->paginate($query->getQuery(), $request->get('page', 1));
    }
}
