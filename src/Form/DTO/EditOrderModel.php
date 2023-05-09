<?php

namespace App\Form\DTO;

use App\Entity\Order;
use Doctrine\Common\Collections\Collection;

/**
 * DTO создана для для избежания конфликта с DateTimeImmutable
 */
class EditOrderModel
{
    public ?int $id = null;

    public ?\DateTime $createdAt = null;

    public ?User $owner = null;

    public ?int $status = null;

    public ?float $totalPrice = null;

    public ?\DateTime $updatedAt = null;

    public ?bool $isDeleted = null;

    public Collection $orderProducts;

    public static function makeFromOrder(?Order $order): self
    {
        $model = new self();

        if (!$order) {
            return $model;
        }

        $model->setId($order->getId());
        $model->setCreatedAt($order->getCreatedAt());
        $model->setOwner($order->getOwner());
        $model->setStatus($order->getStatus());
        $model->setTotalPrice($order->getTotalPrice());
        $model->setUpdatedAt($order->getUpdatedAt());
        $model->setIsDeleted($order->isIsDeleted());
        $model->setOrderProducts($order->getOrderProducts());

        return $model;
    }

    /**
     * @return int|null
     */
    public function getId(): ?int
    {
        return $this->id;
    }

    /**
     * @param int|null $id
     */
    public function setId(?int $id): void
    {
        $this->id = $id;
    }

    /**
     * @return \DateTime|null
     */
    public function getCreatedAt(): ?\DateTime
    {
        return $this->createdAt;
    }

    /**
     * @param \DateTime|null $createdAt
     */
    public function setCreatedAt(?\DateTime $createdAt): void
    {
        $this->createdAt = $createdAt;
    }

    /**
     * @return User|null
     */
    public function getOwner(): ?User
    {
        return $this->owner;
    }

    /**
     * @param User|null $owner
     */
    public function setOwner(?User $owner): void
    {
        $this->owner = $owner;
    }

    /**
     * @return int|null
     */
    public function getStatus(): ?int
    {
        return $this->status;
    }

    /**
     * @param int|null $status
     */
    public function setStatus(?int $status): void
    {
        $this->status = $status;
    }

    /**
     * @return float|null
     */
    public function getTotalPrice(): ?float
    {
        return $this->totalPrice;
    }

    /**
     * @param float|null $totalPrice
     */
    public function setTotalPrice(?float $totalPrice): void
    {
        $this->totalPrice = $totalPrice;
    }

    /**
     * @return \DateTime|null
     */
    public function getUpdatedAt(): ?\DateTime
    {
        return $this->updatedAt;
    }

    /**
     * @param \DateTime|null $updatedAt
     */
    public function setUpdatedAt(?\DateTime $updatedAt): void
    {
        $this->updatedAt = $updatedAt;
    }

    /**
     * @return bool|null
     */
    public function getIsDeleted(): ?bool
    {
        return $this->isDeleted;
    }

    /**
     * @param bool|null $isDeleted
     */
    public function setIsDeleted(?bool $isDeleted): void
    {
        $this->isDeleted = $isDeleted;
    }

    /**
     * @return Collection
     */
    public function getOrderProducts(): Collection
    {
        return $this->orderProducts;
    }

    /**
     * @param Collection $orderProducts
     */
    public function setOrderProducts(Collection $orderProducts): void
    {
        $this->orderProducts = $orderProducts;
    }
}
