<?php

namespace App\Form\DTO;

use App\Entity\Category;

class EditCategoryModel
{
    public ?string $id;

    public string $title;

    /**
     * @param Category|null $category
     * @return static
     */
    public static function makeFromCategory(?Category $category): self
    {
        $model = new self();

        if (!$category) {
            $model->id = null;

            return $model;
        }

        $model->id = $category->getId();
        $model->title = $category->getTitle();

        return $model;
    }



}
